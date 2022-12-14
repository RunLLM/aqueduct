package executor

import (
	exec_env "github.com/aqueducthq/aqueduct/lib/collections/execution_environment"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/engine"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/repos/sqlite"
)

type Repos struct {
	ArtifactRepo       		 repos.Artifact
	ArtifactResultRepo 		 repos.ArtifactResult
	DAGRepo            		 repos.DAG
	DAGEdgeRepo        		 repos.DAGEdge
	DAGResultRepo      		 repos.DAGResult
	ExecutionEnvironmentRepo repos.ExecutionEnvironment
	NotificationRepo   		 repos.Notification
	OperatorRepo       		 repos.Operator
	OperatorResultRepo 		 repos.OperatorResult
	WatcherRepo        		 repos.Watcher
	WorkflowRepo       		 repos.Workflow
}

func createRepos() *Repos {
	return &Repos{
		ArtifactRepo:       	  sqlite.NewArtifactRepo(),
		ArtifactResultRepo: 	  sqlite.NewArtifactResultRepo(),
		DAGRepo:            	  sqlite.NewDAGRepo(),
		DAGEdgeRepo:        	  sqlite.NewDAGEdgeRepo(),
		DAGResultRepo:      	  sqlite.NewDAGResultRepo(),
		ExecutionEnvironmentRepo: sqlite.NewExecutionEnvironmentRepo(),
		NotificationRepo:   	  sqlite.NewNotificationRepo(),
		OperatorRepo:       	  sqlite.NewOperatorRepo(),
		OperatorResultRepo: 	  sqlite.NewOperatorResultRepo(),
		WatcherRepo:        	  sqlite.NewWatcherRepo(),
		WorkflowRepo:       	  sqlite.NewWorklowRepo(),
	}
}

func getEngineRepos(repos *Repos) *engine.Repos {
	return &engine.Repos{
		ArtifactRepo:       	  repos.ArtifactRepo,
		ArtifactResultRepo: 	  repos.ArtifactResultRepo,
		DAGRepo:            	  repos.DAGRepo,
		DAGEdgeRepo:        	  repos.DAGEdgeRepo,
		DAGResultRepo:      	  repos.DAGResultRepo,
		ExecutionEnvironmentRepo: repos.ExecutionEnvironmentRepo,
		NotificationRepo:   	  repos.NotificationRepo,
		OperatorRepo:       	  repos.OperatorRepo,
		OperatorResultRepo: 	  repos.OperatorResultRepo,
		WatcherRepo:        	  repos.WatcherRepo,
		WorkflowRepo:       	  repos.WorkflowRepo,
	}
}
