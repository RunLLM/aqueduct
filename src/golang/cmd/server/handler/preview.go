package handler

import (
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/request"
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
	dag_utils "github.com/aqueducthq/aqueduct/lib/workflow/dag"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/github"
	"github.com/aqueducthq/aqueduct/lib/workflow/orchestrator"
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

type previewArtifactResponse struct {
	SerializationType string `json:"serialization_type"`
	ArtifactType      string `json:"artifact_type"`
	Content           []byte `json:"content"`
}

type previewResponse struct {
	Status          shared.ExecutionStatus                `json:"status"`
	OperatorResults map[uuid.UUID]shared.ExecutionState   `json:"operator_results"`
	ArtifactResults map[uuid.UUID]previewArtifactResponse `json:"artifact_results"`
}

type previewResponseNoArtifacts struct {
	Status          shared.ExecutionStatus              `json:"status"`
	OperatorResults map[uuid.UUID]shared.ExecutionState `json:"operator_results"`
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

func checkError(w http.ResponseWriter, err error) {
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (*PreviewHandler) SendResponse(w http.ResponseWriter, response interface{}) {

	structResponse := response.(previewResponse)

	responseNoArtifacts := previewResponseNoArtifacts{
		Status:          structResponse.Status,
		OperatorResults: structResponse.OperatorResults,
	}

	jsonBlob, err := json.Marshal(responseNoArtifacts)
	checkError(w, err)

	mw := multipart.NewWriter(w)
	w.Header().Set("Content-Type", mw.FormDataContentType())

	fw, errA := mw.CreateFormField("metadata")
	checkError(w, errA)

	_, errB := fw.Write(jsonBlob)
	checkError(w, errB)

	for id, art := range structResponse.ArtifactResults {

		art_content := art.Content
		art_serialization := art.SerializationType
		serialization_type := "serialization_type" + id.String()

		fw, errC := mw.CreateFormField(id.String())
		checkError(w, errC)

		_, errD := fw.Write(art_content)
		checkError(w, errD)

		fw, errE := mw.CreateFormField(serialization_type)
		checkError(w, errE)

		_, errF := fw.Write([]byte(art_serialization))
		checkError(w, errF)
	}

	if err := mw.Close(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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
		true, // this is a preview
		h.Database,
	)
	if err != nil {
		return errorRespPtr, http.StatusInternalServerError, errors.Wrap(err, "Error creating dag object.")
	}

	orch, err := orchestrator.NewAqOrchestrator(
		workflowDag,
		h.JobManager,
		orchestrator.AqueductTimeConfig{
			OperatorPollInterval: previewPollIntervalMillisec,
			ExecTimeout:          orchestrator.DefaultExecutionTimeout,
			CleanupTimeout:       orchestrator.DefaultCleanupTimeout,
		},
		true, // this is a preview
	)
	if err != nil {
		return errorRespPtr, http.StatusInternalServerError, errors.Wrap(err, "Error creating orchestrator.")
	}

	defer orch.Finish(ctx)
	status, err := orch.Execute(ctx)
	if err != nil && err != orchestrator.ErrOpExecSystemFailure && err != orchestrator.ErrOpExecBlockingUserFailure {
		return errorRespPtr, http.StatusInternalServerError, errors.Wrap(err, "Error executing the workflow.")
	}

	statusCode := http.StatusOK
	if err == orchestrator.ErrOpExecSystemFailure {
		statusCode = http.StatusInternalServerError
	} else if err == orchestrator.ErrOpExecBlockingUserFailure {
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
			artifact_metadata, err := artf.GetMetadata(ctx)
			if err != nil {
				return errorRespPtr, http.StatusInternalServerError, err
			}

			content, err := artf.GetContent(ctx)
			if err != nil {
				return errorRespPtr, http.StatusInternalServerError, err
			}
			artifactResults[artf.ID()] = previewArtifactResponse{
				SerializationType: artifact_metadata.SerializationType,
				Content:           content,
				ArtifactType:      artifact_metadata.ArtifactType,
			}
		}
	}

	return previewResponse{
		Status:          status,
		OperatorResults: execStateByOp,
		ArtifactResults: artifactResults,
	}, statusCode, nil
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
