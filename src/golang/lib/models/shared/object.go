package shared

type Object string

const (
	WorkflowObject          Object = "workflow"
	WorkflowDagResultObject Object = "workflow_dag_result"
	OrganizationObject      Object = "organization"
)