package v2

import (
	"context"
	"github.com/aqueducthq/aqueduct/cmd/server/handler"
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
	"net/http"
)

// Route:
//	v2/workflow/{workflowId}/objects
// Method: GET
// Params:
//	`workflowId`: ID for `workflow` object
// Request:
//	Headers:
//		`api-key`: user's API Key
// Response:
//	Body:
//		all objects written by `workflowId`

type WorkflowObjectsGetArgs struct {
	*aq_context.AqContext
	workflowId uuid.UUID
}

type WorkflowObjectsGetResponse struct {
	LoadDetails []views.LoadOperator `json:"object_details"`
}

type WorkflowObjectsGetHandler struct {
	handler.GetHandler

	Database database.Database

	OperatorRepo       repos.Operator
	WorkflowRepo       repos.Workflow
	WorkflowDagRepo    repos.DAG
	ArtifactResultRepo repos.ArtifactResult
}

func (*WorkflowObjectsGetHandler) Name() string {
	return "WorkflowObjectsGet"
}

func (h *WorkflowObjectsGetHandler) Prepare(r *http.Request) (interface{}, int, error) {
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

	return &WorkflowObjectsGetArgs{
		AqContext:  aqContext,
		workflowId: workflowID,
	}, http.StatusOK, nil
}

func (h *WorkflowObjectsGetHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*WorkflowObjectsGetArgs)
	emptyResp := WorkflowObjectsGetResponse{}

	saveOpList, err := GetDistinctLoadOpsByWorkflow(
		ctx,
		args.workflowId,
		h.OperatorRepo,
		h.WorkflowDagRepo,
		h.ArtifactResultRepo,
		h.Database,
	)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, err
	}

	return WorkflowObjectsGetResponse{
		LoadDetails: saveOpList,
	}, http.StatusOK, nil
}

// GetDistinctLoadOpsByWorkflow returns a definitive list of all distinct save operators for a given workflow.
// Fills in any parameterized fields on the operator specs.
func GetDistinctLoadOpsByWorkflow(
	ctx context.Context,
	workflowID uuid.UUID,
	operatorRepo repos.Operator,
	workflowDagRepo repos.DAG,
	artifactResultRepo repos.ArtifactResult,
	db database.Database,
) ([]views.LoadOperator, error) {
	// Get all distinct specs for the workflow.
	saveOpList, err := operatorRepo.GetDistinctLoadOPsByWorkflow(ctx, workflowID, db)
	if err != nil {
		return nil, errors.Wrap(err, "Unexpected error occurred when retrieving workflow objects.")
	}

	// Now, we check the list of save operators for any that have parameterized table names.
	// Since these parameterized names are not present on the operator spec, but are instead filled in on the fly at runtime,
	// we need to fetch those table names from storage. This requires potentially expanding a single save operator into multiple,
	// each of which represents a unique table name that was saved to.
	saveOpIDsToExpand := make([]uuid.UUID, 0, len(saveOpList))
	for _, op := range saveOpList {
		relationalLoadParams, isRelational := connector.CastToRelationalDBLoadParams(op.Spec.Parameters)
		if isRelational {
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
	saveOps, err := operatorRepo.GetNodeBatch(ctx, saveOpIDsToExpand, db)
	if err != nil {
		return nil, errors.Wrap(err, "Unexpected error occurred when fetching parameterized save operators.")
	}

	// Use the save operator's ID as the primary key for organization purposes while fetching the artifacts and
	// then their corresponding artifact results.
	paramArtifactIDBySaveOpID := make(map[uuid.UUID]uuid.UUID, len(saveOps))
	for _, saveOp := range saveOps {
		if len([]uuid.UUID(saveOp.Inputs)) < 2 {
			return nil, errors.Newf("Expected parameterized save operator %s to have multiple inputs!", saveOp.ID)
		}

		// Assumption: The parameters to a save operator always come before the actual artifact to save.
		// There is only one parameter we allow for relational saves.
		tableNameParamArtifactID := saveOp.Inputs[0]
		paramArtifactIDBySaveOpID[saveOp.ID] = tableNameParamArtifactID
	}

	paramArtifactResultsBySavedOpID := make(map[uuid.UUID][]models.ArtifactResult, len(paramArtifactIDBySaveOpID))
	for saveOpID, paramArtifactID := range paramArtifactIDBySaveOpID {
		paramArtifactResults, err := artifactResultRepo.GetByArtifact(ctx, paramArtifactID, db)
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
		dagByArtifactResultID, err := workflowDagRepo.GetByArtifactResultBatch(ctx, paramArtifactResultIDs, db)
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
				log.Warnf("Unable to fetch content for artifact result %s: %v", artifactResultID, err)
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
					log.Warnf("Excluding %s from the returned list of save operators because we could not find any successfully saved table names for it.", saveOp.OperatorID)
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
