package handler

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

// Route: /workflow/{workflowId}/tables
// Method: GET
// Params:
//	`workflowId`: ID for `workflow` object
// Request:
//	Headers:
//		`api-key`: user's API Key
// Response:
//	Body:
//		all tables for the given `workflowId`

type getWorkflowTablesArgs struct {
	*aq_context.AqContext
	workflowId uuid.UUID
}

type getWorkflowTablesResponse struct {
	LoadSpec []connector.Load `json:"table_specs"`
}

type GetWorkflowTablesHandler struct {
	GetHandler

	Database                database.Database
	OperatorReader			operator.Reader
	WorkflowReader         	workflow.Reader
}

func (*GetWorkflowTablesHandler) Name() string {
	return "GetWorkflowTables"
}

func (h *GetWorkflowTablesHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	workflowIdStr := chi.URLParam(r, routes.WorkflowIdUrlParam)
	workflowId, err := uuid.Parse(workflowIdStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed workflow ID.")
	}

	ok, err := h.WorkflowReader.ValidateWorkflowOwnership(
		r.Context(),
		workflowId,
		aqContext.OrganizationId,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error during workflow ownership validation.")
	}
	if !ok {
		return nil, http.StatusBadRequest, errors.Wrap(err, "The organization does not own this workflow.")
	}

	return &getWorkflowTablesArgs{
		AqContext:  aqContext,
		workflowId: workflowId,
	}, http.StatusOK, nil
}

func (h *GetWorkflowTablesHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*getWorkflowTablesArgs)

	emptyResp := getWorkflowTablesResponse{}

	// Get all specs  for the workflow.
	operatorList, err := h.OperatorReader.GetUniqueOperatorsByWorkflowId(ctx, args.workflowId, h.Database)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred when retrieving workflow.")
	}

	loadList := make([]connector.Load, 0, len(operatorList))

	for _, operator := range operatorList {
		if operator.Spec.IsLoad() {
			loadList = append(loadList, *operator.Spec.Load())
		}
	}
	
	// TODO: How dedup
	// TODO: Test by querying
	// create workflow that loads table into db
	// postman (aqueduct start --expose)
	// curl --header "api-key: 1CZR2J96NKXL4UWEO57PID03GAVTHB8Q" http://localhost:8080/api/workflow/9a6dbbb6-a1a6-454b-9a1a-8cf3def16e68/tables
	// {
	// 	"table_specs":[
	// 	   {
	// 		  "service":"SQLite",
	// 		  "integration_id":"95ccb2eb-068c-4a85-853c-3a5dced8f192",
	// 		  "parameters":{
	// 			 "table":"test",
	// 			 "update_mode":"replace"
	// 		  }
	// 	   },
	// 	   {
	// 		  "service":"SQLite",
	// 		  "integration_id":"95ccb2eb-068c-4a85-853c-3a5dced8f192",
	// 		  "parameters":{
	// 			 "table":"test",
	// 			 "update_mode":"replace"
	// 		  }
	// 	   },
	// 	   {
	// 		  "service":"SQLite",
	// 		  "integration_id":"95ccb2eb-068c-4a85-853c-3a5dced8f192",
	// 		  "parameters":{
	// 			 "table":"test",
	// 			 "update_mode":"replace"
	// 		  }
	// 	   },
	// 	   {
	// 		  "service":"SQLite",
	// 		  "integration_id":"95ccb2eb-068c-4a85-853c-3a5dced8f192",
	// 		  "parameters":{
	// 			 "table":"test",
	// 			 "update_mode":"append"
	// 		  }
	// 	   },
	// 	   {
	// 		  "service":"SQLite",
	// 		  "integration_id":"95ccb2eb-068c-4a85-853c-3a5dced8f192",
	// 		  "parameters":{
	// 			 "table":"new_table",
	// 			 "update_mode":"append"
	// 		  }
	// 	   },
	// 	   {
	// 		  "service":"SQLite",
	// 		  "integration_id":"95ccb2eb-068c-4a85-853c-3a5dced8f192",
	// 		  "parameters":{
	// 			 "table":"test",
	// 			 "update_mode":"append"
	// 		  }
	// 	   }
	// 	]
	//  }

// 	curl --header "api-key: 1CZR2J96NKXL4UWEO57PID03GAVTHB8Q" http://localhost:8080/api/workflow/ebbbe4b3-50a7-431f-bbee-a555504ef443/tables
// {"table_specs":[{"service":"SQLite","integration_id":"95ccb2eb-068c-4a85-853c-3a5dced8f192","parameters":{"table":"test","update_mode":"replace"}}]}

	return getWorkflowTablesResponse{
		LoadSpec: loadList,
	}, http.StatusOK, nil
}
