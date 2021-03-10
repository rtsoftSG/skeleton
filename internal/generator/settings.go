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

type RouterChoice string

const (
	GorillaMux RouterChoice = "Gorilla mux"
	GIN        RouterChoice = "GIN"
)

type Settings struct {
	ProjectName          string
	ProjectRootDir       string
	Logger               LoggerChoice
	Database             DBChoice
	Router               RouterChoice
	UseConsul            bool
	SyncConfigWithConsul bool
	UseJaeger            bool
	UsePrometheus        bool

	WithDeps bool
}
