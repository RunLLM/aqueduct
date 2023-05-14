package handler

import (
	"context"
	"net/http"
	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/functional/maps"
	"github.com/aqueducthq/aqueduct/lib/functional/slices"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared/operator/connector"
	"github.com/aqueducthq/aqueduct/lib/models/views"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/storage"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// Route: /workflow/{workflowId}/objects
// Method: GET
// Params:
//	`workflowId`: ID for `workflow` object
// Request:
//	Headers:
//		`api-key`: user's API Key
// Response:
//	Body:
//		all objects written by `workflowId`

type ListWorkflowObjectsArgs struct {
	*aq_context.AqContext
	workflowId uuid.UUID
}

type ListWorkflowObjectsResponse struct {
	LoadDetails []views.LoadOperator `json:"object_details"`
}

type ListWorkflowObjectsHandler struct {
	GetHandler

	Database database.Database

	OperatorRepo       repos.Operator
	WorkflowRepo       repos.Workflow
	WorkflowDagRepo    repos.DAG
	ArtifactResultRepo repos.ArtifactResult
}

func (*ListWorkflowObjectsHandler) Name() string {
	return "ListWorkflowObjects"
}

func (h *ListWorkflowObjectsHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	workflowIDStr := chi.URLParam(r, routes.WorkflowIdUrlParam)
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed workflow ID.")
	}

	ok, err := h.WorkflowRepo.ValidateOrg(
		r.Context(),
		workflowID,
		aqContext.OrgID,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error during workflow ownership validation.")
	}
	if !ok {
		return nil, http.StatusBadRequest, errors.Wrap(err, "The organization does not own this workflow.")
	}

	return &ListWorkflowObjectsArgs{
		AqContext:  aqContext,
		workflowId: workflowID,
	}, http.StatusOK, nil
}

func (h *ListWorkflowObjectsHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*ListWorkflowObjectsArgs)

	emptyResp := ListWorkflowObjectsResponse{}

	// Get all specs for the workflow.
	operatorList, err := h.OperatorRepo.GetDistinctLoadOPsByWorkflow(ctx, args.workflowId, h.Database)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred when retrieving workflow objects.")
	}

	// If there are any parameterized save operators, update the list with any successfully saved table names.
	operatorList, err = h.ExpandOperatorListWithParameterizedTableNames(ctx, operatorList)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, err
	}

	return ListWorkflowObjectsResponse{
		LoadDetails: operatorList,
	}, http.StatusOK, nil
}

// TODO: We look for any loads that were parameterized. In such cases, we have to go and fetch their corresponding parameter
// input artifact results and update the table name accordingly.
func (h *ListWorkflowObjectsHandler) ExpandOperatorListWithParameterizedTableNames(
	ctx context.Context,
	saveOpList []views.LoadOperator,
) ([]views.LoadOperator, error) {
	saveOpIDsToExpand := make([]uuid.UUID, 0, len(saveOpList))
	for _, op := range saveOpList {
		if relationalLoadParams := op.Spec.Parameters.(*connector.GenericRelationalDBLoadParams); relationalLoadParams != nil {
			if relationalLoadParams.Table == "" {
				saveOpIDsToExpand = append(saveOpIDsToExpand, op.OperatorID)
			}
		}
	}

	// If there are no parameterized saved table names, continue without expanding the list.
	if len(saveOpIDsToExpand) == 0 {
		return saveOpList, nil
	}

	// Fetch each of the load operators	that need to be expanded.
	saveOps, err := h.OperatorRepo.GetNodeBatch(ctx, saveOpIDsToExpand, h.Database)
	if err != nil {
		return nil, errors.Wrap(err, "Unexpected error occurred when fetching parameterized save operators.")
	}

	// Use the save operator's ID as the primary key for organization when fetching the artifacts and then their corresponding
	// artifact results.
	paramArtifactIDBySaveOpID := make(map[uuid.UUID]uuid.UUID, len(saveOps))
	for _, saveOp := range saveOps {
		if len([]uuid.UUID(saveOp.Inputs)) < 2 {
			return nil, errors.Newf("Expected parameterized save operator %s to have multiple inputs!", saveOp.ID)
		}

		tableNameParamArtifactID := saveOp.Inputs[0]
		paramArtifactIDBySaveOpID[saveOp.ID] = tableNameParamArtifactID
	}

	paramArtifactResultsBySavedOpID := make(map[uuid.UUID][]models.ArtifactResult, len(paramArtifactIDBySaveOpID))
	for saveOpID, paramArtifactID := range paramArtifactIDBySaveOpID {
		paramArtifactResults, err := h.ArtifactResultRepo.GetByArtifact(ctx, paramArtifactID, h.Database)
		if err != nil {
			return nil, errors.Wrapf(err, "Unable to fetch artifact result for artifact %s", paramArtifactID)
		}
		paramArtifactResultsBySavedOpID[saveOpID] = append(paramArtifactResultsBySavedOpID[saveOpID], paramArtifactResults...)
	}

	// Fetch the table name for each individual artifact result. If it doesn't exist, we simply skip the artifact result.
	tableNameByParamArtifactResultID := make(map[uuid.UUID]string, len(paramArtifactResultsBySavedOpID))
	for _, paramArtifactResults := range paramArtifactResultsBySavedOpID {
		paramArtifactResultIDs := slices.Map(
			paramArtifactResults,
			func(artifactResult models.ArtifactResult) uuid.UUID {
				return artifactResult.ID
			},
		)
		dagByArtifactResultID, err := h.WorkflowDagRepo.GetByArtifactResultBatch(ctx, paramArtifactResultIDs, h.Database)
		if err != nil {
			return nil, errors.Wrap(err, "Unexpected error when fetching DAGs from artifact results.")
		}

		paramArtifactResultByID := maps.FromValues(
			paramArtifactResults,
			func(artifactResult models.ArtifactResult) uuid.UUID {
				return artifactResult.ID
			},
		)
		for artifactResultID, dag := range dagByArtifactResultID {
			storageConfig := dag.StorageConfig
			storageObj := storage.NewStorage(&storageConfig)

			// We perform a best-effort fetch of the artifact results. If the table name parameter was never executed,
			// we simply continue onwards - perhaps the execution had failed.
			contentBytes, err := storageObj.Get(ctx, paramArtifactResultByID[artifactResultID].ContentPath)
			if err != nil {
				log.Warn("Unable to fetch content for artifact result %s: %v", artifactResultID, err)
				continue
			}
			tableNameByParamArtifactResultID[artifactResultID] = string(contentBytes)
		}
	}

	expandedOpList := make([]views.LoadOperator, 0, len(saveOpList))
	for _, saveOp := range saveOpList {
		if paramArtifactResults, ok := paramArtifactResultsBySavedOpID[saveOp.OperatorID]; ok {

			// Grab all the table names for each save operator.
			tableNames := make(map[string]bool, len(paramArtifactResults))
			for _, paramArtifactResult := range paramArtifactResults {
				if tableName, ok := tableNameByParamArtifactResultID[paramArtifactResult.ID]; ok {
					tableNames[tableName] = true
				} else {
					// No table names were found for this save operator, so let's skip the operator it altogether.
					log.Warn("Excluding %s from the returned list of save operators because we could not find any successfully saved table names for it.", saveOp.OperatorID)
					continue
				}
			}

			// Perform the expansion of the save operator into multiple, one for each unique table name.
			// Keep all the non-parameter parts of the spec the same though.
			for tableName := range tableNames {
				// This performs a shallow copy, meaning everything except the parameters is correctly copied over.
				newSave := saveOp

				saveParams, isRelational := connector.CastToRelationalDBLoadParams(newSave.Spec.Parameters)
				if !isRelational {
					return nil, errors.Wrapf(err, "Unexpected error when casting load parameters for operator %s to relational DB load parameters.", saveOp.OperatorID)
				}
				newSave.Spec.Parameters = &connector.GenericRelationalDBLoadParams{
					RelationalDBLoadParams: connector.RelationalDBLoadParams{
						Table:      tableName,
						UpdateMode: saveParams.UpdateMode,
					},
				}
				expandedOpList = append(expandedOpList, newSave)
			}
		} else {
			// This operator does not need to be expanded. Leave it as is.
			expandedOpList = append(expandedOpList, saveOp)
		}
	}
	return expandedOpList, nil
}
