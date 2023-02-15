package function

import "github.com/aqueducthq/aqueduct/lib/models/shared/operator/connector/github"

type Function struct {
	Type           Type                   `json:"type"`
	Language       string                 `json:"language"`
	Granularity    Granularity            `json:"granularity"`
	StoragePath    string                 `json:"storage_path"`
	GithubMetadata *github.GithubMetadata `json:"github_metadata,omitempty"`
	EntryPoint     *EntryPoint            `json:"entry_point,omitempty"`
	CustomArgs     string                 `json:"custom_args"`
}

type Type string

const (
	FileFunctionType    Type = "file"
	GithubFunctionType  Type = "github"
	BuiltInFunctionType Type = "built_in"
)

type Granularity string

const (
	TableGranularity Granularity = "table"
	RowGranularity   Granularity = "row"
)

type EntryPoint struct {
	File      string `json:"file"`
	ClassName string `json:"class_name"`
	Method    string `json:"method"`
}
