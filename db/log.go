package db

// Logging keys for slog contextual fields
const (
	LogDatabase  = "database" // the database key
	LogSql       = "query"
	LogArgs      = "args"
	LogError     = "error"
	LogComponent = "component"
	LogTable     = "table"
	LogColumn    = "column"
	LogFilename  = "filename"
	LogStartTime = "start"
	LogEndTime   = "end"
	LogDuration  = "duration"
)
