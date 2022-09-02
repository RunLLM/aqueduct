package handler

import (
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"

	artifact_db "github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/cmd/server/request"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact_result"
	"github.com/aqueducthq/aqueduct/lib/collections/integration"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/engine"
	dag_utils "github.com/aqueducthq/aqueduct/lib/workflow/dag"
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

type previewArgs struct {
	*aq_context.AqContext
	DagSummary *request.DagSummary
	// Add list of IDs
}

type previewResponse struct {
	Status          shared.ExecutionStatus                `json:"status"`
	OperatorResults map[uuid.UUID]shared.ExecutionState   `json:"operator_results"`
	ArtifactContents map[uuid.UUID][]byte 				  `json:"artifact_contents"`
	ArtifactTypesMetadata map[uuid.UUID]artifactTypeMetadata		  `json:"artifact_types_metadata"`
}

type previewResponseMetadata struct {
	Status          shared.ExecutionStatus                `json:"status"`
	OperatorResults map[uuid.UUID]shared.ExecutionState   `json:"operator_results"`
	ArtifactTypesMetadata map[uuid.UUID]artifactTypeMetadata		  `json:"artifact_types_metadata"`
}

type artifactTypeMetadata struct {
	SerializationType artifact_result.SerializationType `json:"serialization_type"`
	ArtifactType      artifact_db.Type                  `json:"artifact_type"`
}


type PreviewHandler struct {
	PostHandler

	Database          database.Database
	IntegrationReader integration.Reader
	GithubManager     github.Manager
	AqEngine          engine.AqEngine
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

	resp := response.(*previewResponse)
	mw := multipart.NewWriter(w)
	w.Header().Set("Content-Type", mw.FormDataContentType())

	responseMetadata := previewResponseMetadata{
		Status: resp.Status,
		OperatorResults: resp.OperatorResults,
		ArtifactTypesMetadata: resp.ArtifactTypesMetadata,
	}

	jsonBlob, err := json.Marshal(responseMetadata)
	checkError(w, err)

	fw, err := mw.CreateFormField("metadata")
	checkError(w, err)

	_, err = fw.Write(jsonBlob)
	checkError(w, err)

	for artifact_id, artifact_content := range resp.ArtifactContents {
		fw, err := mw.CreateFormField(artifact_id.String())
		checkError(w, err)

		_, err = fw.Write(artifact_content)
		checkError(w, err)
	}

	err = mw.Close()
	checkError(w, err)
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
		aqContext.StorageConfig,
	)
	if err != nil {
		return nil, statusCode, err
	}

	ok, err := dag_utils.ValidateDagOperatorIntegrationOwnership(
		r.Context(),
		dagSummary.Dag.Operators,
		aqContext.OrganizationId,
		aqContext.Id,
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

	timeConfig := &engine.AqueductTimeConfig{
		OperatorPollInterval: engine.DefaultPollIntervalMillisec,
		ExecTimeout:          engine.DefaultExecutionTimeout,
		CleanupTimeout:       engine.DefaultCleanupTimeout,
	}

	workflowPreviewResult, err := h.AqEngine.PreviewWorkflow(
		ctx,
		dagSummary.Dag,
		timeConfig,
	)
	if err != nil && err != engine.ErrOpExecSystemFailure && err != engine.ErrOpExecBlockingUserFailure {
		return errorRespPtr, http.StatusInternalServerError, errors.Wrap(err, "Error executing the workflow.")
	}

	statusCode := http.StatusOK
	if err == engine.ErrOpExecSystemFailure {
		statusCode = http.StatusInternalServerError
	} else if err == engine.ErrOpExecBlockingUserFailure {
		statusCode = http.StatusBadRequest
	}


	// Only include artifact results that were successfully computed.
	artifactContents := make(map[uuid.UUID][]byte)
	artifactTypesMetadata := make(map[uuid.UUID]artifactTypeMetadata)
	for id, artf := range workflowPreviewResult.Artifacts {
		artifactContents[id] = artf.Content
		artifactTypesMetadata[id] = artifactTypeMetadata{
			SerializationType: artf.SerializationType,
			ArtifactType: artf.ArtifactType,
		}
	}

	return &previewResponse{
		Status:          workflowPreviewResult.Status,
		OperatorResults: workflowPreviewResult.Operators,
		ArtifactContents: artifactContents,
		ArtifactTypesMetadata: artifactTypesMetadata,
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
