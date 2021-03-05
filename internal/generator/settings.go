package generator

type LoggerChoice string

const (
	GoKit LoggerChoice = "Go Kit"
	Zap   LoggerChoice = "Zap"
)

type DBChoice string

const (
	NoDb       DBChoice = "No database"
	Clickhouse DBChoice = "Clickhouse"
	Postgresql DBChoice = "Postgres"
)

type Settings struct {
	ProjectName          string
	ProjectRootDir       string
	Logger               LoggerChoice
	Database             DBChoice
	UseConsul            bool
	SyncConfigWithConsul bool
	UseJaeger            bool
}
