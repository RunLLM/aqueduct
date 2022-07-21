package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/aqueducthq/aqueduct/cmd/server/request"
	db_artifact "github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact_result"
	"github.com/aqueducthq/aqueduct/lib/collections/integration"
	"github.com/aqueducthq/aqueduct/lib/collections/notification"
	"github.com/aqueducthq/aqueduct/lib/collections/operator_result"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/collections/user"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_result"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/artifact"
	dag_utils "github.com/aqueducthq/aqueduct/lib/workflow/dag"
	"github.com/aqueducthq/aqueduct/lib/workflow/engine"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/github"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

// Route: /preview
// Method: POST
// Params: none
// Request:
//	Headers:
//		`api-key`: user's API Key
//	Body:
//		"dag": a serialized `workflow_dag` object
//		"<operator_id>": zip file associated with operator for the `operator_id`.
//  	"<operator_id>": ... (more operator files)
// Response:
//	Body:
//		serialized `previewResponse` object consisting of overall status and results for all executed operators / artifacts.

const previewPollIntervalMillisec = 100

type previewArgs struct {
	*aq_context.AqContext
	DagSummary *request.DagSummary
	// Add list of IDs
}

type previewFloatArtifactResponse struct {
	Val float64 `json:"val"`
}

type previewBoolArtifactResponse struct {
	Passed bool `json:"passed"`
}

type previewParamArtifactResponse struct {
	Val string `json:"val"`
}

type previewTableArtifactResponse struct {
	TableSchema []map[string]string `json:"table_schema"`
	Data        string              `json:"data"`
}

type previewArtifactResponse struct {
	Table  *previewTableArtifactResponse `json:"table"`
	Metric *previewFloatArtifactResponse `json:"metric"`
	Check  *previewBoolArtifactResponse  `json:"check"`
	Param  *previewParamArtifactResponse `json:"param"`
}

type previewResponse struct {
	Status          shared.ExecutionStatus                `json:"status"`
	OperatorResults map[uuid.UUID]shared.ExecutionState   `json:"operator_results"`
	ArtifactResults map[uuid.UUID]previewArtifactResponse `json:"artifact_results"`
}

type PreviewHandler struct {
	PostHandler

	Database          database.Database
	IntegrationReader integration.Reader
	StorageConfig     *shared.StorageConfig
	JobManager        job.JobManager
	GithubManager     github.Manager
	Vault             vault.Vault
}

func (*PreviewHandler) Name() string {
	return "Preview"
}

func (h *PreviewHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	dagSummary, statusCode, err := request.ParseDagSummaryFromRequest(
		r,
		aqContext.Id,
		h.GithubManager,
		h.StorageConfig,
	)
	if err != nil {
		return nil, statusCode, err
	}

	ok, err := dag_utils.ValidateDagOperatorIntegrationOwnership(
		r.Context(),
		dagSummary.Dag.Operators,
		aqContext.OrganizationId,
		h.IntegrationReader,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error during integration ownership validation.")
	}
	if !ok {
		return nil, http.StatusBadRequest, errors.Wrap(err, "The organization does not own the integrations defined in the Dag.")
	}

	removeLoadOperators(dagSummary)

	if err := dag_utils.Validate(
		dagSummary.Dag,
	); err != nil {
		if _, ok := dag_utils.ValidationErrors[err]; !ok {
			return nil, http.StatusInternalServerError, errors.Wrap(err, "Internal system error occurred while validating the DAG.")
		} else {
			return nil, http.StatusBadRequest, err
		}
	}

	return &previewArgs{
		AqContext:  aqContext,
		DagSummary: dagSummary,
	}, http.StatusOK, nil
}

func (h *PreviewHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*previewArgs)
	errorRespPtr := &previewResponse{Status: shared.FailedExecutionStatus}
	dagSummary := args.DagSummary

	_, err := operator.UploadOperatorFiles(ctx, dagSummary.Dag, dagSummary.FileContentsByOperatorUUID)
	if err != nil {
		return errorRespPtr, http.StatusInternalServerError, errors.Wrap(err, "Error uploading function files.")
	}

	workflowDag, err := dag_utils.NewWorkflowDag(
		ctx,
		dagSummary.Dag,
		workflow_dag_result.NewNoopWriter(true),
		operator_result.NewNoopWriter(true),
		artifact_result.NewNoopWriter(true),
		workflow.NewNoopReader(true),
		notification.NewNoopWriter(true),
		user.NewNoopReader(true),
		h.JobManager,
		h.Vault,
		h.StorageConfig,
		h.Database,
	)
	if err != nil {
		return errorRespPtr, http.StatusInternalServerError, errors.Wrap(err, "Error creating dag object.")
	}

	eng, err := engine.NewAqEngine(
		workflowDag,
		h.Database,
		h.GithubManager,
		h.Vault,
		h.JobManager,
		engine.AqueductTimeConfig{
			OperatorPollInterval: previewPollIntervalMillisec,
			ExecTimeout:          engine.DefaultExecutionTimeout,
			CleanupTimeout:       engine.DefaultCleanupTimeout,
		},
		false, /* shouldPersistResults */
	)
	if err != nil {
		return errorRespPtr, http.StatusInternalServerError, errors.Wrap(err, "Error creating orchestrator.")
	}

	defer eng.Finish(ctx)
	status, err := eng.Execute(ctx)
	if err != nil && err != engine.ErrOpExecSystemFailure && err != engine.ErrOpExecBlockingUserFailure {
		return errorRespPtr, http.StatusInternalServerError, errors.Wrap(err, "Error executing the workflow.")
	}

	statusCode := http.StatusOK
	if err == engine.ErrOpExecSystemFailure {
		statusCode = http.StatusInternalServerError
	} else if err == engine.ErrOpExecBlockingUserFailure {
		statusCode = http.StatusBadRequest
	}

	execStateByOp := make(map[uuid.UUID]shared.ExecutionState, len(workflowDag.Operators()))
	for _, op := range workflowDag.Operators() {
		execState, err := op.GetExecState(ctx)
		if err != nil {
			return errorRespPtr, http.StatusInternalServerError, err
		}
		execStateByOp[op.ID()] = *execState
	}

	// Only include artifact results that were successfully computed.
	artifactResults := make(map[uuid.UUID]previewArtifactResponse)
	for _, artf := range workflowDag.Artifacts() {
		if artf.Computed(ctx) {
			artifactResp, err := convertToPreviewArtifactResponse(ctx, artf)
			if err != nil {
				return errorRespPtr, http.StatusInternalServerError, err
			}
			artifactResults[artf.ID()] = *artifactResp
		}
	}

	return &previewResponse{
		Status:          status,
		OperatorResults: execStateByOp,
		ArtifactResults: artifactResults,
	}, statusCode, nil
}

func convertToPreviewArtifactResponse(ctx context.Context, artf artifact.Artifact) (*previewArtifactResponse, error) {
	content, err := artf.GetContent(ctx)
	if err != nil {
		return nil, err
	}

	if artf.Type() == db_artifact.FloatType {
		val, err := strconv.ParseFloat(string(content), 32)
		if err != nil {
			return nil, err
		}

		return &previewArtifactResponse{
			Metric: &previewFloatArtifactResponse{
				Val: val,
			},
		}, nil
	} else if artf.Type() == db_artifact.BoolType {
		passed, err := strconv.ParseBool(string(content))
		if err != nil {
			return nil, err
		}

		return &previewArtifactResponse{
			Check: &previewBoolArtifactResponse{
				Passed: passed,
			},
		}, nil
	} else if artf.Type() == db_artifact.JsonType {
		return &previewArtifactResponse{
			Param: &previewParamArtifactResponse{
				Val: string(content),
			},
		}, nil
	} else if artf.Type() == db_artifact.TableType {
		metadata, err := artf.GetMetadata(ctx)
		if err != nil {
			metadata = &artifact_result.Metadata{}
		}
		return &previewArtifactResponse{
			Table: &previewTableArtifactResponse{
				TableSchema: metadata.Schema,
				Data:        string(content),
			},
		}, nil
	}
	return nil, errors.Newf("Unsupported artifact type %s", artf.Type())
}

func removeLoadOperators(dagSummary *request.DagSummary) {
	removeList := make([]uuid.UUID, 0, len(dagSummary.Dag.Operators))

	for id, op := range dagSummary.Dag.Operators {
		if op.Spec.IsLoad() {
			removeList = append(removeList, id)
		}
	}

	for _, id := range removeList {
		delete(dagSummary.Dag.Operators, id)
	}
}
