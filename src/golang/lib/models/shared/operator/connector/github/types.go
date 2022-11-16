package github

import (
	"fmt"
)

const RepoConfigPath = ".aqconfig"

type GithubExtractType string

const (
	ZipExtractType    GithubExtractType = "zip"
	StringExtractType GithubExtractType = "string"
)

type GithubRepoContentType string

const (
	FileGithubRepoContentType GithubRepoContentType = "file"
	DirGithubRepoContentType  GithubRepoContentType = "dir"
)

type GithubRepoConfigContentType string

const (
	OperatorGithubRepoConfigContentType GithubRepoConfigContentType = "operator"
	QueryGithubRepoConfigContentType    GithubRepoConfigContentType = "query"
)

type GithubMetadata struct {
	Owner  string `json:"owner"`
	Repo   string `json:"repo"`
	Branch string `json:"branch"`
	Path   string `json:"path"`

	// If `RepoConfigContentType` and `RepoConfigContentName` are specified,
	// we will retrieve content based on .aqueduct.config instead of `Path`.
	// `Path` field will be backfilled based on updated repo config.
	RepoConfigContentType GithubRepoConfigContentType `json:"repo_config_content_type,omitempty"`
	RepoConfigContentName string                      `json:"repo_config_content_name,omitempty"`

	// Repo config obtained from `.aqueduct.config`.
	RepoConfig *RepoConfig `json:"repo_config,omitempty"`
	// The commit ID corresponding to the exact GH version.
	// Typically, used to determine if the branch has been updated and if we need to
	// re-pull the branch
	CommitId string `json:"commit_id"`
}

func (g GithubMetadata) RepoUrl() string {
	return fmt.Sprintf("%s/%s", g.Owner, g.Repo)
}

type OperatorRepoConfig struct {
	Path       string `yaml:"path",json:"path"`
	EntryPoint string `yaml:"entry_point",json:"entry_point"`
	ClassName  string `yaml:"class_name",json:"class_name"`
	Method     string `yaml:"method",json:"method"`
}

type QueryRepoConfig struct {
	Path string `yaml:"path",json:"path"`
}

type RepoConfig struct {
	Operators map[string]OperatorRepoConfig `yaml:"operators",json:"operators"`
	Queries   map[string]QueryRepoConfig    `yaml:"queries",json:"queries"`
}
