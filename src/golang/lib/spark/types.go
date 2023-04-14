package spark

type StatementState string

type SessionState string

type StatementOutputStatus string

const (
	NotStarted   SessionState = "not_started"
	Starting     SessionState = "starting"
	Idle         SessionState = "idle"
	Busy         SessionState = "busy"
	ShuttingDown SessionState = "shutting_down"
	SessionError SessionState = "error"
	Dead         SessionState = "dead"
	Killed       SessionState = "killed"
	Success      SessionState = "success"

	Waiting        StatementState = "waiting"
	Running        StatementState = "running"
	Available      StatementState = "available"
	StatementError StatementState = "error"
	Cancelling     StatementState = "cancelling"

	OK    StatementOutputStatus = "ok"
	Error StatementOutputStatus = "error"
)

// Livy Session.
type Session struct {
	ID        int          `json:"id"`
	AppID     string       `json:"appId,omitempty"`
	Owner     string       `json:"owner,omitempty"`
	ProxyUser string       `json:"proxyUser,omitempty"`
	Kind      string       `json:"kind,omitempty"`
	Log       []string     `json:"log,omitempty"`
	State     SessionState `json:"state"`
	AppInfo   AppInfo      `json:"appInfo,omitempty"`
}

type AppInfo struct {
	DriverLogUrl string `json:"driverLogUrl"`
	SparkUiUrl   string `json:"sparkUiUrl"`
}

type Log struct {
	Log string `json:"log"`
	ID  int    `json:"id"`
}

type Conf struct {
	SparkExecutorEnv []string `json:"spark.executorEnv"`
	SparkDriverEnv   []string `json:"spark.driverEnv"`
}

type Batch struct {
	ID      int               `json:"id"`
	AppID   string            `json:"appId"`
	AppInfo map[string]string `json:"appInfo"`
	Log     []string          `json:"log"`
	State   string            `json:"state"`
}

type Statement struct {
	ID        int             `json:"id"`
	Code      string          `json:"code"`
	State     StatementState  `json:"state"`
	Output    StatementOutput `json:"output"`
	Progress  float64         `json:"progress"`
	Started   int64           `json:"started"`
	Completed int64           `json:"completed"`
}

type StatementOutput struct {
	Status StatementOutputStatus `json:"status"`
	// Although "execution_count" doesn't follow the convention, it is correct
	// https://livy.incubator.apache.org/docs/latest/rest-api.html
	ExecutionCount int                    `json:"execution_count"`
	Data           map[string]interface{} `json:"data"`
}

// BatchRequest represents the request body for creating a batch
type BatchRequest struct {
	File      string            `json:"file,omitempty"`
	ClassName string            `json:"className,omitempty"`
	Args      []string          `json:"args,omitempty"`
	Conf      map[string]string `json:"conf,omitempty"`
	ProxyUser string            `json:"proxyUser,omitempty"`
	Files     []string          `json:"files,omitempty"`
	Jars      []string          `json:"jars,omitempty"`
	PyFiles   []string          `json:"pyFiles,omitempty"`
	Code      string            `json:"code,omitempty"`
}

// StatementRequest represents the request body for creating a statement
type StatementRequest struct {
	Code string `json:"code"`
}

// CreateSessionRequest is request body for creating a session
type CreateSessionRequest struct {
	Kind                     string            `json:"kind"`
	ProxyUser                string            `json:"proxyUser,omitempty"`
	Jars                     []string          `json:"jars,omitempty"`
	PyFiles                  []string          `json:"pyFiles,omitempty"`
	Files                    []string          `json:"files,omitempty"`
	DriverMemory             string            `json:"driverMemory,omitempty"`
	DriverCores              int               `json:"driverCores,omitempty"`
	ExecutorMemory           string            `json:"executorMemory,omitempty"`
	ExecutorCores            int               `json:"executorCores,omitempty"`
	NumExecutors             int               `json:"numExecutors,omitempty"`
	Archives                 []string          `json:"archives,omitempty"`
	Queue                    string            `json:"queue,omitempty"`
	Name                     string            `json:"name,omitempty"`
	Conf                     map[string]string `json:"conf,omitempty"`
	HeartbeatTimeoutInSecond int               `json:"heartbeatTimeoutInSecond,omitempty"`
}

type GetSessionsResponse struct {
	From     int       `json:"from"`
	Total    int       `json:"total"`
	Sessions []Session `json:"sessions"`
}
