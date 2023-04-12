package routes

const (
	WorkflowIdUrlParam          = "workflowId"
	WorkflowDagIDUrlParam       = "workflowDagID"
	WorkflowDagResultIdUrlParam = "workflowDagResultId"
	OperatorIdUrlParam          = "operatorId"
	ArtifactIdUrlParam          = "artifactId"
	NotificationIdUrlParam      = "notificationId"
	IntegrationIdUrlParam       = "integrationId"

	// v2 params
	// Each V2 parameters should have a corresponding parser
	// in request/parser package.
	WorkflowIDParam   = "workflowID"
	DagIDParam        = "dagID"
	DAGResultIDParam  = "dagResultID"
	NodeIDParam       = "nodeID"
	NodeResultIDParam = "nodeResultID"
)
