package usage

const (
	schemaVersion int    = 1
	obfuscated    string = "***"
	delimiter     string = "/"
	hashKey       string = "aqueduct"
	// This is the Grafana Loki server address
	logURL string = "https://355951:eyJrIjoiYjgyNTFhMjk4NTcxOGViYjk0MDRmYjM4OTdmMDZlNWNmZmM1MmI1ZCIsIm4iOiJhcXVlZHVjdF8wIiwiaWQiOjc3MDAzMH0=@logs-prod-017.grafana.net/loki/api/v1/push"

	// Env vars
	codespacesEnv            string = "CODESPACES"
	codespacesEnvActiveValue string = "true"

	codespaceNameEnv string = "CODESPACE_NAME"
)
