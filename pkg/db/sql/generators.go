package sql

import (
	"fmt"
	"github.com/goradd/iter"
	"github.com/goradd/orm/pkg/db/jointree"
	. "github.com/goradd/orm/pkg/query"
	"strings"
)

type operationSqler interface {
	OperationSql(op Operator, operandStrings []string) string
}

type deleteUsesAliaser interface {
	DeleteUsesAlias() bool
}

// selectGenerator is an aid to generating various sql statements.
// SQL dialects are similar, but have small variations. This object
// attempts to handle the major issues, while allowing individual
// implementations of SQL to do their own tweaks.
type selectGenerator struct {
	jt      *jointree.JoinTree
	dbi     DbI
	argList []any
}

func newSelectGenerator(jt *jointree.JoinTree, dbi DbI) *selectGenerator {
	return &selectGenerator{jt: jt, dbi: dbi}
}

func (g *selectGenerator) iq(v string) string {
	return g.dbi.QuoteIdentifier(v)
}

func (g *selectGenerator) addArg(v any) string {
	g.argList = append(g.argList, v)
	return g.dbi.FormatArgument(len(g.argList))
}

func (g *selectGenerator) generateSelectSql() (sql string) {
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

	return sb.String()
}

func (g *selectGenerator) generateDeleteSql() (sql string) {
	var sb strings.Builder

	if t, ok := g.dbi.(deleteUsesAliaser); ok && t.DeleteUsesAlias() {
		j := g.jt.Root
		alias := g.iq(j.Alias)
		sb.WriteString("DELETE ")
		sb.WriteString(alias)
		sb.WriteString("\n")
	} else {
		sb.WriteString("DELETE\n")
	}

	sb.WriteString(g.generateFromSql())
	sb.WriteString(g.generateWhereSql())
	sb.WriteString(g.generateOrderBySql())
	sb.WriteString(g.generateLimitSql())

	return sb.String()
}

func (g *selectGenerator) generateColumnListWithAliases() (sql string) {
	var sb strings.Builder

	// Iterate over root selects and append to the string builder
	for e := range g.jt.Root.SelectsIter() {
		sb.WriteString(g.generateColumnNodeSql(e.Parent.Alias, e.QueryNode))
		sb.WriteString(" AS ")
		sb.WriteString(g.iq(e.Alias))
		sb.WriteString(",\n")
	}

	// Sort to keep the resulting query predictable
	for _, v := range iter.KeySort(g.jt.Aliases) {
		node := v.(Node)
		aliaser := v.(Aliaser)
		sb.WriteString(g.generateNodeSql(node, false))
		alias := aliaser.Alias()
		if alias != "" {
			// This happens in a subquery
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

func (g *selectGenerator) generateColumnNodeSql(parentAlias string, node Node) (sql string) {
	var sb strings.Builder

	sb.WriteString(g.iq(parentAlias))
	sb.WriteString(".")
	sb.WriteString(g.iq(node.(*ColumnNode).QueryName))

	return sb.String()
}

func (g *selectGenerator) generateNodeSql(n Node, useAlias bool) (sql string) {
	switch node := n.(type) {
	case *ValueNode:
		v := ValueNodeGetValue(node)
		if a, ok := v.([]Node); ok {
			// value is actually a list of nodes
			var l []string
			for _, o := range a {
				l = append(l, g.generateNodeSql(o, useAlias))
			}
			return strings.Join(l, ",")
		} else {
			return g.addArg(v)
		}
	case *OperationNode:
		return g.generateOperationSql(node, useAlias)
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
			n := g.jt.Aliases[node.Alias()]
			if n != nil {
				sql = g.generateNodeSql(n, false)
			}
		}

	case *SubqueryNode:
		sql = g.generateSubquerySql(node)
	case TableNodeI:
		tj := g.jt.FindElement(node)
		sql = g.generateColumnNodeSql(tj.Alias, node.PrimaryKey())
	default:
		panic("Can't generate sql from node type.")
	}
	return
}

func (g *selectGenerator) generateOperationSql(n *OperationNode, useAlias bool) (sql string) {
	if useAlias && n.Alias() != "" {
		return g.iq(n.Alias())
	}

	var sb strings.Builder
	var operands []string
	operator := OperationNodeOperator(n)

	for _, o := range OperationNodeOperands(n) {
		operands = append(operands, g.generateNodeSql(o, useAlias))
	}

	if o, ok := g.dbi.(operationSqler); ok {
		sql = o.OperationSql(operator, operands)
		if sql != "" {
			return sql
		}
	}

	switch operator {
	case OpFunc:
		if len(operands) > 0 {
			sb.WriteString(strings.Join(operands, ","))
		} else if OperationNodeFunction(n) == "COUNT" {
			sb.WriteString("*")
		}

		if OperationNodeDistinct(n) {
			content := sb.String()
			sb.Reset()
			sb.WriteString("DISTINCT ")
			sb.WriteString(content)
		}
		sb.WriteString(OperationNodeFunction(n))
		sb.WriteString("(")
		sb.WriteString(sb.String())
		sb.WriteString(") ")

	case OpNull, OpNotNull:
		s := operands[0]
		sb.WriteString("(")
		sb.WriteString(s)
		sb.WriteString(" IS ")
		sb.WriteString(operator.String())
		sb.WriteString(") ")

	case OpNot:
		s := operands[0]
		sb.WriteString("(")
		sb.WriteString(operator.String())
		sb.WriteString(" ")
		sb.WriteString(s)
		sb.WriteString(") ")

	case OpIn, OpNotIn:
		s := operands[0]
		sb.WriteString(s)
		sb.WriteString(" ")
		sb.WriteString(operator.String())
		sb.WriteString(" (")
		sb.WriteString(operands[1])
		sb.WriteString(") ")

	case OpAll, OpNone:
		sb.WriteString("(")
		sb.WriteString(operator.String())
		sb.WriteString(") ")

	case OpStartsWith:
		s := g.argList[len(g.argList)-1].(string)
		s += "%"
		g.argList[len(g.argList)-1] = s
		sb.WriteString("(")
		sb.WriteString(fmt.Sprintf(`%s LIKE %s`, operands[0], operands[1]))
		sb.WriteString(")")

	case OpEndsWith:
		s := g.argList[len(g.argList)-1].(string)
		s = "%" + s
		g.argList[len(g.argList)-1] = s
		sb.WriteString("(")
		sb.WriteString(fmt.Sprintf(`%s LIKE %s`, operands[0], operands[1]))
		sb.WriteString(")")

	case OpContains:
		s := g.argList[len(g.argList)-1].(string)
		s = "%" + s + "%"
		g.argList[len(g.argList)-1] = s
		sb.WriteString("(")
		sb.WriteString(fmt.Sprintf(`%s LIKE %s`, operands[0], operands[1]))
		sb.WriteString(")")

	case OpDateAddSeconds:
		panic("DateAddSeconds is not implemented in this database")

	case OpXor:
		s := operands[0]
		s2 := operands[1]
		sb.WriteString(fmt.Sprintf(`(((%[1]s) AND NOT (%[2]s)) OR (NOT (%[1]s) AND (%[2]s)))`, s, s2))

	default:
		sOp := " " + operator.String() + " "
		sb.WriteString(" (")
		sb.WriteString(strings.Join(operands, sOp))
		sb.WriteString(") ")
	}

	return sb.String()
}

func (g *selectGenerator) generateAlias(alias string) (sql string) {
	return g.iq(alias)
}

func (g *selectGenerator) generateSubquerySql(node *SubqueryNode) (sql string) {
	// The copy below intentionally reuses the argList and db items
	g2 := *g
	//g2.b = SubqueryBuilder(node).(*jointree.Builder)
	sql = g2.generateSelectSql()
	sql = "(" + sql + ")"
	return
}

func (g *selectGenerator) generateFromSql() (sql string) {
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

func (g *selectGenerator) generateJoinSql(j *jointree.Element) (sql string) {
	var sb strings.Builder

	var tn TableNodeI
	var ok bool

	if tn, ok = j.QueryNode.(TableNodeI); !ok {
		panic("cannot generate join code for a non-table node")
	}

	switch tn.NodeType_() {
	case ReferenceNodeType:
		ref := tn.(ReferenceNodeI)
		sb.WriteString("LEFT JOIN ")
		sb.WriteString(g.iq(tn.TableName_()))
		sb.WriteString(" AS ")
		sb.WriteString(g.iq(j.Alias))
		sb.WriteString(" ON ")
		sb.WriteString(g.iq(j.Parent.Alias))
		sb.WriteString(".")
		sb.WriteString(g.iq(ref.ColumnName()))
		sb.WriteString(" = ")
		sb.WriteString(g.iq(j.Alias))
		sb.WriteString(".")
		sb.WriteString(g.iq(ref.PrimaryKey().QueryName))
		if j.JoinCondition != nil {
			s := g.generateNodeSql(j.JoinCondition, false)
			sb.WriteString(" AND ")
			sb.WriteString(s)
		}
	case ReverseNodeType:
		rev := tn.(ReverseNodeI)
		if g.jt.Limits.AreSet() && !rev.IsExpanded() {
			panic("We do not currently support limited queries with an array join.")
		}

		sb.WriteString("LEFT JOIN ")
		sb.WriteString(g.iq(rev.TableName_()))
		sb.WriteString(" AS ")
		sb.WriteString(g.iq(j.Alias))
		sb.WriteString(" ON ")
		sb.WriteString(g.iq(j.Parent.Alias))
		sb.WriteString(".")
		sb.WriteString(g.iq(j.Parent.QueryNode.(PrimaryKeyer).PrimaryKey().QueryName))
		sb.WriteString(" = ")
		sb.WriteString(g.iq(j.Alias))
		sb.WriteString(".")
		sb.WriteString(g.iq(rev.ColumnName()))
		if j.JoinCondition != nil {
			s := g.generateNodeSql(j.JoinCondition, false)
			sb.WriteString(" AND ")
			sb.WriteString(s)
		}
	case ManyManyNodeType:
		mm := tn.(ManyManyNodeI)

		if g.jt.Limits.AreSet() && !mm.IsExpanded() {
			panic("We do not currently support limited queries with an array join.")
		}

		sb.WriteString("LEFT JOIN ")
		sb.WriteString(g.iq(mm.AssnTableName()))
		sb.WriteString(" AS ")
		sb.WriteString(g.iq(j.Alias + "a"))
		sb.WriteString(" ON ")
		sb.WriteString(g.iq(j.Parent.Alias))
		sb.WriteString(".")
		sb.WriteString(g.iq(NodeParent(mm).(TableNodeI).PrimaryKey().QueryName))
		sb.WriteString(" = ")
		sb.WriteString(g.iq(j.Alias + "a"))
		sb.WriteString(".")
		sb.WriteString(g.iq(mm.RefColumnName()))
		sb.WriteString("\n")

		sb.WriteString("LEFT JOIN ")
		sb.WriteString(g.iq(mm.TableName_()))
		sb.WriteString(" AS ")
		sb.WriteString(g.iq(j.Alias))
		sb.WriteString(" ON ")
		sb.WriteString(g.iq(j.Alias + "a"))
		sb.WriteString(".")
		sb.WriteString(g.iq(mm.RefColumnName()))
		sb.WriteString(" = ")
		sb.WriteString(g.iq(j.Alias))
		sb.WriteString(".")
		sb.WriteString(g.iq(mm.PrimaryKey().QueryName))
		if j.JoinCondition != nil {
			s := g.generateNodeSql(j.JoinCondition, false)
			sb.WriteString(" AND ")
			sb.WriteString(s)
		}
	default:
		return
	}

	sb.WriteString("\n")
	for _, cj := range j.References {
		sb.WriteString(g.generateJoinSql(cj))
	}

	return sb.String()
}

func (g *selectGenerator) generateWhereSql() (sql string) {
	if g.jt.Condition != nil {
		var sb strings.Builder
		sb.WriteString("WHERE ")
		sb.WriteString(g.generateNodeSql(g.jt.Condition, false))
		sb.WriteString("\n")
		return sb.String()
	}
	return
}

func (g *selectGenerator) generateGroupBySql() (sql string) {
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

func (g *selectGenerator) generateHaving() (sql string) {
	if g.jt.Having != nil {
		var sb strings.Builder
		sb.WriteString("HAVING ")
		sb.WriteString(g.generateNodeSql(g.jt.Having, false))
		sb.WriteString("\n")
		return sb.String()
	}
	return
}

func (g *selectGenerator) generateLimitSql() (sql string) {
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

func (g *selectGenerator) generateOrderBySql() (sql string) {
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
func GenerateUpdate(db DbI, table string, fields map[string]any, where map[string]any) (sql string, args []any) {
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

	for k, v := range iter.KeySort(where) {
		args = append(args, v)
		sb.WriteString(db.QuoteIdentifier(k))
		sb.WriteString("=")
		sb.WriteString(db.FormatArgument(len(args)))
		sb.WriteString(" AND ")
	}

	// Remove trailing " AND "
	sql = strings.TrimSuffix(sb.String(), " AND ")

	return sql, args
}

// GenerateInsert is a helper function for database implementations to generate an insert statement.
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
func GenerateDelete(db DbI, table string, where map[string]any) (sql string, args []any) {
	var sb strings.Builder

	sb.WriteString("DELETE FROM ")
	sb.WriteString(db.QuoteIdentifier(table))
	sb.WriteString("\nWHERE ")

	var s string
	s, args = generateWhereClause(db, where)
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
		s, args = generateWhereClause(db, where)
		sb.WriteString(s)
	}

	if len(orderBy) > 0 {
		sb.WriteString("\nORDER BY ")
		sb.WriteString(strings.Join(orderBy, ", "))
	}

	return sb.String(), args
}

func generateWhereClause(db DbI, where map[string]any) (sql string, args []any) {
	var ors []string
	for kOr, vOr := range iter.KeySort(where) {
		if m, ok := vOr.(map[string]any); ok {
			var ands []string
			for kAnd, vAnd := range m {
				args = append(args, vAnd)
				var sb strings.Builder
				sb.WriteString(db.QuoteIdentifier(kAnd))
				sb.WriteString("=")
				sb.WriteString(db.FormatArgument(len(args)))
				ands = append(ands, sb.String())
			}
			and := strings.Join(ands, " AND ")
			ors = append(ors, and)
		} else {
			args = append(args, vOr)
			var sb strings.Builder
			sb.WriteString(db.QuoteIdentifier(kOr))
			sb.WriteString("=")
			sb.WriteString(db.FormatArgument(len(args)))
			ors = append(ors, sb.String())
		}
	}
	sql = strings.Join(ors, " OR ")
	return
}
