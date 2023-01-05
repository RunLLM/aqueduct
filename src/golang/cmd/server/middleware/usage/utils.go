package usage

type Labels struct {
	SchemaVersion int    `json:"schema_version"`
	Environment   string `json:"environment"`
	StatusCode    int    `json:"status_code"`
	Route         string `json:"route"`
}

type Payload struct {
	ID      string `json:"id"`
	Latency int64  `json:"latency"`
	Labels
}

// These fields and json label names for Stream and Streams are required by Loki.
type Stream struct {
	Labels Labels     `json:"stream"`
	Values [][]string `json:"values"`
}

type Streams struct {
	Streams []Stream `json:"streams"`
}
