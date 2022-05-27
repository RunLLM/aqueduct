package tests

import (
	"context"
	"testing"
)

// resetDatabase wipes all rows from the database
func resetDatabase(t *testing.T) {
	resetWorkflowDagResult(t)
	resetWorkflowDagEdge(t)
	resetOperator(t)
	resetWorkflowDag(t)
	resetWorkflow(t)
	resetIntegration(t)
	resetUser(t)
}

func resetUser(t *testing.T) {
	if err := db.Execute(context.Background(), "DELETE FROM app_user;"); err != nil {
		t.Errorf("Unable to reset app_user table: %v", err)
		t.FailNow()
	}
}

func resetIntegration(t *testing.T) {
	if err := db.Execute(context.Background(), "DELETE FROM integration;"); err != nil {
		t.Errorf("Unable to reset integration table: %v", err)
		t.FailNow()
	}
}

func resetWorkflow(t *testing.T) {
	if err := db.Execute(context.Background(), "DELETE FROM workflow;"); err != nil {
		t.Errorf("Unable to reset workflow table: %v", err)
		t.FailNow()
	}
}

func resetWorkflowDag(t *testing.T) {
	if err := db.Execute(context.Background(), "DELETE FROM workflow_dag;"); err != nil {
		t.Errorf("Unable to reset workflow_dag table: %v", err)
		t.FailNow()
	}
}

func resetOperator(t *testing.T) {
	if err := db.Execute(context.Background(), "DELETE FROM operator;"); err != nil {
		t.Errorf("Unable to reset operator table: %v", err)
		t.FailNow()
	}
}

func resetWorkflowDagEdge(t *testing.T) {
	if err := db.Execute(context.Background(), "DELETE FROM workflow_dag_edge;"); err != nil {
		t.Errorf("Unable to reset workflow_dag_edge table: %v", err)
		t.FailNow()
	}
}

func resetWorkflowDagResult(t *testing.T) {
	if err := db.Execute(context.Background(), "DELETE FROM workflow_dag_result;"); err != nil {
		t.Errorf("Unable to reset workflow_dag_result table: %v", err)
		t.FailNow()
	}
}
