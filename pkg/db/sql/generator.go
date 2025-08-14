package sql

import (
	"fmt"
	"strings"

	"github.com/goradd/anyutil"
	"github.com/goradd/iter"
	"github.com/goradd/orm/pkg/db/jointree"
	. "github.com/goradd/orm/pkg/query"
)

type operationSqler interface {
	OperationSql(op Operator, operands []Node, operandStrings []string) string
}

type deleteUsesAliaser interface {
	DeleteUsesAlias() bool
}

type forUpdater interface {
	ForUpdate() bool
}

// sqlGenerator is an aid to generating various sql statements.
// SQL dialects are similar, but have small variations. This object
// attempts to handle the major issues, while allowing individual
// implementations of SQL to do their own tweaks.
type sqlGenerator struct {
	jt      *jointree.JoinTree
	dbi     DbI
	argList []any
}

func newSqlGenerator(jt *jointree.JoinTree, dbi DbI) *sqlGenerator {
	return &sqlGenerator{jt: jt, dbi: dbi}
}

// iq quotes an identifier in the way the current SQL dialect accepts.
func (g *sqlGenerator) iq(v string) string {
	return g.dbi.QuoteIdentifier(v)
}

func (g *sqlGenerator) addArg(v any) string {
	g.argList = append(g.argList, v)
	return g.dbi.FormatArgument(len(g.argList))
}

func (g *sqlGenerator) generateSelectSql() (sql string, args []any) {
	var sb strings.Builder

	if g.jt.IsDistinct {
		sb.WriteString("SELECT DISTINCT\n")
	} else {
		sb.WriteString("SELECT\n")
	}

	sb.WriteString(g.generateColumnListWithAliases())
	sb.WriteString(g.generateFromSql())
	sb.WriteString(g.generateWhereSql())
	sb.WriteString(g.generateGroupBySql())
	sb.WriteString(g.generateHaving())
	sb.WriteString(g.generateOrderBySql())
	sb.WriteString(g.generateLimitSql())

	return sb.String(), g.argList
}

func (g *sqlGenerator) generateColumnListWithAliases() (sql string) {
	var sb strings.Builder

	// Iterate over root selects and append to the string builder
	for e := range g.jt.SelectsIter() {
		sb.WriteString(g.generateColumnNodeSql(e.Parent.Alias, e.QueryNode))
		sb.WriteString(" AS ")
		sb.WriteString(g.iq(e.Alias))
		sb.WriteString(",\n")
	}

	// Sort to keep the resulting query predictable
	for alias, node := range g.jt.CalculationsIter() {
		sb.WriteString(g.generateNodeSql(node, false))
		if alias != "" {
			sb.WriteString(" AS ")
			sb.WriteString(g.iq(alias))
		}
		sb.WriteString(",\n")
	}

	// Convert the builder to a string and trim the trailing comma and newline
	sql = strings.TrimSuffix(sb.String(), ",\n")
	sql += "\n"
	return
}

func (g *sqlGenerator) generateColumnNodeSql(parentAlias string, node Node) (sql string) {
	var sb strings.Builder

	sb.WriteString(g.iq(parentAlias))
	sb.WriteString(".")
	sb.WriteString(g.iq(ColumnNodeQueryName(node)))

	return sb.String()
}

// generateNodeSql will create the sql for a node. If the node is a collection of nodes, it will generate it for the collection separated by commas.
func (g *sqlGenerator) generateNodeSql(n Node, useAlias bool) (sql string) {
	switch node := n.(type) {
	case *ValueNode:
		v := ValueNodeGetValue(node)
		if a, ok := v.([]Node); ok {
			// value is actually a list of nodes
			var l []string
			for _, o := range a {
				l = append(l, g.generateNodeSql(o, useAlias))
			}
			sql = strings.Join(l, ",")
		} else {
			sql = g.addArg(v)
		}
	case *OperationNode:
		sql = g.generateOperationSql(node, useAlias)
	case *ColumnNode:
		item := g.jt.FindElement(node)
		if useAlias && item.Alias != "" {
			sql = g.generateAlias(item.Alias)
		} else {
			sql = g.generateColumnNodeSql(item.Parent.Alias, node)
		}
	case *AliasNode:
		if useAlias {
			sql = g.iq(node.Alias())
		} else {
			n := g.jt.FindAlias(node.Alias())
			if n != nil {
				sql = g.generateNodeSql(n, false)
			}
		}

	case *SubqueryNode:
		//sql = g.generateSubquerySql(node)
	case TableNodeI:
		if len(node.PrimaryKeys()) > 1 {
			panic("cannot use a table node for a table with a composite key as a value")
		}
		tj := g.jt.FindElement(node)
		pkNode := node.PrimaryKeys()[0]
		sql = g.generateColumnNodeSql(tj.Alias, pkNode)
	default:
		panic("Can't generate sql from node type.")
	}
	return
}

// generateOperationSql generates SQL for an operation node.
// useAlias specifies whether the operands can be aliased or not.
func (g *sqlGenerator) generateOperationSql(n *OperationNode, useAlias bool) (sql string) {
	var sb strings.Builder
	var operandStrings []string
	var operands []Node

	operator := OperationNodeOperator(n)
	operands = OperationNodeOperands(n)

	for _, o := range operands {
		operandStrings = append(operandStrings, g.generateNodeSql(o, useAlias))
	}

	if o, ok := g.dbi.(operationSqler); ok {
		if sql = o.OperationSql(operator, operands, operandStrings); sql != "" {
			return
		}
	}

	switch operator {
	case OpFunc:
		sb.WriteString(OperationNodeFunction(n))
		sb.WriteString("(")
		if OperationNodeDistinct(n) {
			sb.WriteString("DISTINCT ")
		}

		if len(operandStrings) > 0 {
			sb.WriteString(strings.Join(operandStrings, ","))
		} else if OperationNodeFunction(n) == "COUNT" {
			sb.WriteString("*")
		}
		sb.WriteString(") ")

	case OpNull, OpNotNull:
		s := operandStrings[0]
		sb.WriteString("(")
		sb.WriteString(s)
		sb.WriteString(" IS ")
		sb.WriteString(operator.String())
		sb.WriteString(") ")

	case OpNot:
		s := operandStrings[0]
		sb.WriteString("(")
		sb.WriteString(operator.String())
		sb.WriteString(" ")
		sb.WriteString(s)
		sb.WriteString(") ")

	case OpIn, OpNotIn:
		s := operandStrings[0]
		sb.WriteString(s)
		sb.WriteString(" ")
		sb.WriteString(operator.String())
		sb.WriteString(" (")
		sb.WriteString(operandStrings[1])
		sb.WriteString(") ")

	case OpAll, OpNone:
		sb.WriteString("(")
		sb.WriteString(operator.String())
		sb.WriteString(") ")

	case OpStartsWith:
		s := fmt.Sprint(g.argList[len(g.argList)-1])
		s += "%"
		g.argList[len(g.argList)-1] = s
		sb.WriteString("(")
		sb.WriteString(fmt.Sprintf(`%s LIKE %s`, operandStrings[0], operandStrings[1]))
		sb.WriteString(")")

	case OpEndsWith:
		s := fmt.Sprint(g.argList[len(g.argList)-1])
		s = "%" + s
		g.argList[len(g.argList)-1] = s
		sb.WriteString("(")
		sb.WriteString(fmt.Sprintf(`%s LIKE %s`, operandStrings[0], operandStrings[1]))
		sb.WriteString(")")

	case OpContains:
		s := fmt.Sprint(g.argList[len(g.argList)-1])
		s = "%" + s + "%"
		g.argList[len(g.argList)-1] = s
		sb.WriteString("(")
		sb.WriteString(fmt.Sprintf(`%s LIKE %s`, operandStrings[0], operandStrings[1]))
		sb.WriteString(")")

	case OpDateAddSeconds:
		panic("DateAddSeconds is not implemented in this database")

	case OpXor:
		s := operandStrings[0]
		s2 := operandStrings[1]
		sb.WriteString(fmt.Sprintf(`(((%[1]s) AND NOT (%[2]s)) OR (NOT (%[1]s) AND (%[2]s)))`, s, s2))

	default:
		sOp := " " + operator.String() + " "
		sb.WriteString(" (")
		sb.WriteString(strings.Join(operandStrings, sOp))
		sb.WriteString(") ")
	}

	return sb.String()
}

func (g *sqlGenerator) generateAlias(alias string) (sql string) {
	return g.iq(alias)
}

func (g *sqlGenerator) generateCountSql() (sql string, args []any) {
	if g.jt.HasSelects() || g.jt.HasCalcs() {
		// Use a subquery to get the rows, then just count the rows
		sql, args = g.generateSelectSql()

		sql = "SELECT COUNT(*) FROM (" + sql + ") AS s"
		return
	}
	// No need to subquery. Just count on the query.
	var sb strings.Builder

	sb.WriteString("SELECT COUNT(*)\n")
	sb.WriteString(g.generateFromSql())
	sb.WriteString(g.generateWhereSql())
	sb.WriteString(g.generateGroupBySql())
	sb.WriteString(g.generateHaving())

	return sb.String(), g.argList
}

func (g *sqlGenerator) generateFromSql() (sql string) {
	var sb strings.Builder

	sb.WriteString("FROM\n")

	j := g.jt.Root
	sb.WriteString(g.iq(j.QueryNode.TableName_()))
	sb.WriteString(" AS ")
	sb.WriteString(g.iq(j.Alias))
	sb.WriteString("\n")

	for _, child := range j.References {
		sb.WriteString(g.generateJoinSql(child))
	}

	return sb.String()
}

func (g *sqlGenerator) generateJoinSql(j *jointree.Element) (sql string) {
	var sb strings.Builder

	var tn TableNodeI
	var ok bool

	if tn, ok = j.QueryNode.(TableNodeI); !ok {
		panic("cannot generate join code for a non-table node")
	}

	switch tn.NodeType_() {
	case ReferenceNodeType:
		ref := tn.(ReferenceNodeI)
		fk, pk := ref.ColumnNames()
		sb.WriteString("LEFT JOIN ")
		sb.WriteString(g.iq(tn.TableName_()))
		sb.WriteString(" AS ")
		sb.WriteString(g.iq(j.Alias))
		sb.WriteString(" ON ")
		sb.WriteString(g.iq(j.Parent.Alias))
		sb.WriteString(".")
		sb.WriteString(g.iq(fk))
		sb.WriteString(" = ")
		sb.WriteString(g.iq(j.Alias))
		sb.WriteString(".")
		sb.WriteString(g.iq(pk))
	case ReverseNodeType:
		rev := tn.(ReverseNodeI)
		fk, pk := rev.ColumnNames()

		if g.jt.Limits.AreSet() {
			panic("We do not currently support limited queries with an array join.")
		}

		sb.WriteString("LEFT JOIN ")
		sb.WriteString(g.iq(rev.TableName_()))
		sb.WriteString(" AS ")
		sb.WriteString(g.iq(j.Alias))
		sb.WriteString(" ON ")
		sb.WriteString(g.iq(j.Parent.Alias))
		sb.WriteString(".")
		sb.WriteString(g.iq(pk))
		sb.WriteString(" = ")
		sb.WriteString(g.iq(j.Alias))
		sb.WriteString(".")
		sb.WriteString(g.iq(fk))
	case ManyManyNodeType:
		mm := tn.(ManyManyNodeI)
		fkp, pkp := mm.ParentColumnNames()
		fkr, pkr := mm.RefColumnNames()

		if g.jt.Limits.AreSet() {
			panic("We do not currently support limited queries with an array join.")
		}

		sb.WriteString("LEFT JOIN ")
		sb.WriteString(g.iq(mm.AssnTableName()))
		sb.WriteString(" AS ")
		sb.WriteString(g.iq(j.Alias + "a"))
		sb.WriteString(" ON ")
		sb.WriteString(g.iq(j.Parent.Alias))
		sb.WriteString(".")
		sb.WriteString(g.iq(pkp))
		sb.WriteString(" = ")
		sb.WriteString(g.iq(j.Alias + "a"))
		sb.WriteString(".")
		sb.WriteString(g.iq(fkp))
		sb.WriteString("\n")

		sb.WriteString("LEFT JOIN ")
		sb.WriteString(g.iq(mm.TableName_()))
		sb.WriteString(" AS ")
		sb.WriteString(g.iq(j.Alias))
		sb.WriteString(" ON ")
		sb.WriteString(g.iq(j.Alias + "a"))
		sb.WriteString(".")
		sb.WriteString(g.iq(fkr))
		sb.WriteString(" = ")
		sb.WriteString(g.iq(j.Alias))
		sb.WriteString(".")
		sb.WriteString(g.iq(pkr))
	default:
		return
	}

	sb.WriteString("\n")
	for _, cj := range j.References {
		sb.WriteString(g.generateJoinSql(cj))
	}

	return sb.String()
}

func (g *sqlGenerator) generateWhereSql() (sql string) {
	if g.jt.Condition != nil {
		var sb strings.Builder
		sb.WriteString("WHERE ")
		sb.WriteString(g.generateNodeSql(g.jt.Condition, false))
		sb.WriteString("\n")
		return sb.String()
	}
	return
}

func (g *sqlGenerator) generateGroupBySql() (sql string) {
	if len(g.jt.GroupBys) > 0 {
		var sb strings.Builder
		sb.WriteString("GROUP BY ")
		for _, n := range g.jt.GroupBys {
			sb.WriteString(g.generateNodeSql(n, true))
			sb.WriteString(",")
		}
		sql := sb.String()
		sql = strings.TrimSuffix(sql, ",")
		sql += "\n"
		return sql
	}
	return
}

func (g *sqlGenerator) generateHaving() (sql string) {
	if g.jt.Having != nil {
		var sb strings.Builder
		sb.WriteString("HAVING ")
		sb.WriteString(g.generateNodeSql(g.jt.Having, false))
		sb.WriteString("\n")
		return sb.String()
	}
	return
}

func (g *sqlGenerator) generateLimitSql() (sql string) {
	if !g.jt.Limits.AreSet() {
		return ""
	}

	var sb strings.Builder

	if g.jt.Limits.MaxRowCount > 0 {
		sb.WriteString(fmt.Sprintf("LIMIT %d ", g.jt.Limits.MaxRowCount))
	}

	if g.jt.Limits.Offset > 0 {
		sb.WriteString(fmt.Sprintf("OFFSET %d ", g.jt.Limits.Offset))
	}

	return sb.String()
}

func (g *sqlGenerator) generateOrderBySql() (sql string) {
	if len(g.jt.OrderBys) > 0 {
		var sb strings.Builder
		sb.WriteString("ORDER BY ")
		for _, n := range g.jt.OrderBys {
			sb.WriteString(g.generateNodeSql(n, true))
			if n.IsDescending() {
				sb.WriteString(" DESC")
			}
			sb.WriteString(",")
		}
		sql := sb.String()
		sql = strings.TrimSuffix(sql, ",")
		sql += "\n"
		return sql
	}
	return
}

// GenerateUpdate is a helper function for database implementations to generate an update statement.
// useOr will indicate whether to OR or AND the items in the where group. If where has a map[string]any object in it,
// the items in that map will be OR'd or AND'd opposite to userOr. This is recursive.
func GenerateUpdate(db DbI, table string, fields map[string]any, where map[string]any, useOr bool) (sql string, args []any) {
	if len(fields) == 0 {
		panic("No fields to set")
	}

	var sb strings.Builder

	sb.WriteString("UPDATE ")
	sb.WriteString(db.QuoteIdentifier(table))
	sb.WriteString("\nSET ")

	// We range on sorted keys to give SQL optimizers a chance to use a prepared
	// statement by making sure the same fields show up in the same order.
	for k, v := range iter.KeySort(fields) {
		args = append(args, v)
		sb.WriteString(db.QuoteIdentifier(k))
		sb.WriteString("=")
		sb.WriteString(db.FormatArgument(len(args)))
		sb.WriteString(", ")
	}

	// Remove trailing ", "
	sql = strings.TrimSuffix(sb.String(), ", ")
	sb.Reset()
	sb.WriteString(sql)
	sb.WriteString("\nWHERE ")

	var s2 string
	s2, args = generateWhereClause(db, where, useOr, args)
	sb.WriteString(s2)
	sql = sb.String()

	return
}

// GenerateInsert is a helper function for database implementations to generate an insert statement.
// Pass the quoted table name to quotedTable. Include the schema name here if applicaable.
func GenerateInsert(db DbI, table string, fields map[string]any) (sql string, args []any) {
	if len(fields) == 0 {
		panic("No fields to insert")
	}

	var sb strings.Builder

	sb.WriteString("INSERT INTO ")
	sb.WriteString(db.QuoteIdentifier(table))
	sb.WriteString(" (")

	var keys []string
	var values []string

	// We range on sorted keys to give SQL optimizers a chance to use a prepared
	// statement by making sure the same fields show up in the same order.
	for k, v := range iter.KeySort(fields) {
		keys = append(keys, db.QuoteIdentifier(k))
		args = append(args, v)
		values = append(values, db.FormatArgument(len(args)))
	}

	sb.WriteString(strings.Join(keys, ","))
	sb.WriteString(")\nVALUES (")
	sb.WriteString(strings.Join(values, ","))
	sb.WriteString(")\n")

	return sb.String(), args
}

// GenerateDelete is a helper function for database implementations to generate a delete statement.
// useOr will determine if the where items are initially ANDed or ORed
func GenerateDelete(db DbI, table string, where map[string]any, useOr bool) (sql string, args []any) {
	var sb strings.Builder

	sb.WriteString("DELETE FROM ")
	sb.WriteString(db.QuoteIdentifier(table))
	if where == nil {
		return sb.String(), nil // delete everything
	}
	if len(where) == 0 {
		panic("An empty where map cannot be provided") // Prevent a dangerous programming mistake that would wipe out an entire table.
	}
	sb.WriteString("\nWHERE ")

	var s string
	s, args = generateWhereClause(db, where, useOr, args)
	sb.WriteString(s)

	return sb.String(), args
}

// GenerateSelect is a helper function for database implementations to generate a select statement.
func GenerateSelect(db DbI, table string, fieldNames []string, where map[string]any, orderBy []string) (sql string, args []any) {
	if len(fieldNames) == 0 {
		panic("No fields to select")
	}

	var sb strings.Builder

	sb.WriteString("SELECT ")
	sb.WriteString(strings.Join(fieldNames, ",\n"))
	sb.WriteString("\nFROM ")
	sb.WriteString(db.QuoteIdentifier(table))

	if len(where) > 0 {
		sb.WriteString("\nWHERE ")
		var s string
		s, args = generateWhereClause(db, where, true, args)
		sb.WriteString(s)
	}

	if len(orderBy) > 0 {
		sb.WriteString("\nORDER BY ")
		sb.WriteString(strings.Join(orderBy, ", "))
	}

	return sb.String(), args
}

// GenerateVersionLock is a helper function for database implementations to implement optimistic locking.
// It generates the sql that will return the version number of the record, while also doing an exclusive write
// lock on the row, if a transaction is active. If a transaction is not active, it will simply do a read of the version number.
func GenerateVersionLock(db DbI, table string, pkName string, pkValue any, versionName string, inTransaction bool) (sql string, args []any) {
	var sb strings.Builder

	sb.WriteString("SELECT ")
	sb.WriteString(db.QuoteIdentifier(versionName))
	sb.WriteString("\nFROM ")
	sb.WriteString(db.QuoteIdentifier(table))
	sb.WriteString("\nWHERE ")
	sb.WriteString(db.QuoteIdentifier(pkName))
	sb.WriteString(" = ")
	args = append(args, pkValue)
	sb.WriteString(db.FormatArgument(len(args)))

	if inTransaction && db.SupportsForUpdate() {
		sb.WriteString(" FOR UPDATE")
	}
	return sb.String(), args
}

func generateWhereClause(db DbI, where map[string]any, connectWithOr bool, argsIn []any) (sql string, argsOut []any) {
	var clauses []string
	argsOut = argsIn
	for key, value := range iter.KeySort(where) {
		if m, ok := value.(map[string]any); ok {
			var sql2 string
			sql2, argsOut = generateWhereClause(db, m, !connectWithOr, argsOut)
			clauses = append(clauses, sql2)
		} else if ints, ok2 := value.([]int); ok2 {
			s2 := db.QuoteIdentifier(key)
			s2 += " IN ("
			s2 += anyutil.Join(ints, ",")
			s2 += ")"
			clauses = append(clauses, s2)
		} else if strs, ok3 := value.([]string); ok3 {
			var formattedStrings []string

			for _, s := range strs {
				argsOut = append(argsOut, s)
				formattedStrings = append(formattedStrings, db.FormatArgument(len(argsOut)))
			}

			s2 := db.QuoteIdentifier(key)
			s2 += " IN ("
			s2 += strings.Join(formattedStrings, ",")
			s2 += ")"
			clauses = append(clauses, s2)
		} else {
			argsOut = append(argsOut, value)
			var sb strings.Builder
			sb.WriteString(db.QuoteIdentifier(key))
			sb.WriteString("=")
			sb.WriteString(db.FormatArgument(len(argsOut)))
			clauses = append(clauses, sb.String())
		}
	}
	var sep string
	if connectWithOr {
		sep = " OR "
	} else {
		sep = " AND "
	}
	sql = "(" + strings.Join(clauses, sep) + ")"
	return
}
