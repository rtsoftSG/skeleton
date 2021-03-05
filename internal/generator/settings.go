package generator

type LoggerChoice string

const (
	GoKit LoggerChoice = "Go Kit"
	Zap   LoggerChoice = "Zap"
)

type DBChoice string

const (
	Clickhouse DBChoice = "Clickhouse"
	Postgresql DBChoice = "Postgres"
)

type Settings struct {
	ProjectName    string
	ProjectRootDir string
	Logger         LoggerChoice
	Database       DBChoice
	Consul         bool
}
