package handler

import (
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/request"
	artifact_db "github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact_result"
	db_exec_env "github.com/aqueducthq/aqueduct/lib/collections/execution_environment"
	"github.com/aqueducthq/aqueduct/lib/collections/integration"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/engine"
	exec_env "github.com/aqueducthq/aqueduct/lib/execution_environment"
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
	Status                shared.ExecutionStatus              `json:"status"`
	OperatorResults       map[uuid.UUID]shared.ExecutionState `json:"operator_results"`
	ArtifactContents      map[uuid.UUID][]byte                `json:"artifact_contents"`
	ArtifactTypesMetadata map[uuid.UUID]artifactTypeMetadata  `json:"artifact_types_metadata"`
}

type previewResponseMetadata struct {
	Status                shared.ExecutionStatus              `json:"status"`
	OperatorResults       map[uuid.UUID]shared.ExecutionState `json:"operator_results"`
	ArtifactTypesMetadata map[uuid.UUID]artifactTypeMetadata  `json:"artifact_types_metadata"`
}

type artifactTypeMetadata struct {
	SerializationType artifact_result.SerializationType `json:"serialization_type"`
	ArtifactType      artifact_db.Type                  `json:"artifact_type"`
}

type PreviewHandler struct {
	PostHandler

	Database                   database.Database
	IntegrationReader          integration.Reader
	ExecutionEnvironmentReader db_exec_env.Reader
	ExecutionEnvironmentWriter db_exec_env.Writer
	GithubManager              github.Manager
	AqEngine                   engine.AqEngine
}

func (*PreviewHandler) Name() string {
	return "Preview"
}

// This custom implementation of SendResponse constructs a multipart form response with the following fields:
// "metadata" contains a json serialized blob of operator and artifact result metadata.
// For each artifact, it generates a field with artifact id as the field name and artifact content
// as the value.
func (*PreviewHandler) SendResponse(w http.ResponseWriter, response interface{}) {
	resp := response.(*previewResponse)
	multipartWriter := multipart.NewWriter(w)
	defer multipartWriter.Close()

	w.Header().Set("Content-Type", multipartWriter.FormDataContentType())

	responseMetadata := previewResponseMetadata{
		Status:                resp.Status,
		OperatorResults:       resp.OperatorResults,
		ArtifactTypesMetadata: resp.ArtifactTypesMetadata,
	}

	jsonBlob, err := json.Marshal(responseMetadata)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fw, err := multipartWriter.CreateFormField(metadataFormFieldName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = fw.Write(jsonBlob)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for artifact_id, artifact_content := range resp.ArtifactContents {
		fw, err := multipartWriter.CreateFormField(artifact_id.String())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = fw.Write(artifact_content)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
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

	execEnvByOpId, err := h.setupExecEnv(ctx, args)
	if err != nil {
		return errorRespPtr, http.StatusInternalServerError, errors.Wrap(err, "Error setting up python environments.")
	}

	timeConfig := &engine.AqueductTimeConfig{
		OperatorPollInterval: engine.DefaultPollIntervalMillisec,
		ExecTimeout:          engine.DefaultExecutionTimeout,
		CleanupTimeout:       engine.DefaultCleanupTimeout,
	}

	workflowPreviewResult, err := h.AqEngine.PreviewWorkflow(
		ctx,
		dagSummary.Dag,
		execEnvByOpId,
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

	artifactContents := make(map[uuid.UUID][]byte, len(workflowPreviewResult.Artifacts))
	artifactTypesMetadata := make(map[uuid.UUID]artifactTypeMetadata, len(workflowPreviewResult.Artifacts))
	for id, artf := range workflowPreviewResult.Artifacts {
		artifactContents[id] = artf.Content
		artifactTypesMetadata[id] = artifactTypeMetadata{
			SerializationType: artf.SerializationType,
			ArtifactType:      artf.ArtifactType,
		}
	}

	return &previewResponse{
		Status:                workflowPreviewResult.Status,
		OperatorResults:       workflowPreviewResult.Operators,
		ArtifactContents:      artifactContents,
		ArtifactTypesMetadata: artifactTypesMetadata,
	}, statusCode, nil
}

func (h *PreviewHandler) setupExecEnv(
	ctx context.Context,
	args *previewArgs,
) (map[uuid.UUID]exec_env.ExecutionEnvironment, error) {
	condaConnected, err := exec_env.IsCondaConnected(
		ctx, args.Id, h.IntegrationReader, h.Database,
	)

	dagSummary := args.DagSummary

	if err != nil {
		return nil, errors.Wrap(err, "Unable to verify if conda is connected.")
	}

	// For now, do nothing if conda is not connected.
	if !condaConnected {
		return nil, nil
	}

	rawEnvByOperator := make(
		map[uuid.UUID]exec_env.ExecutionEnvironment,
		len(dagSummary.FileContentsByOperatorUUID),
	)

	for opId, zipball := range dagSummary.FileContentsByOperatorUUID {
		rawEnv, err := exec_env.InferDependenciesFromZipFile(zipball)
		if err != nil {
			return nil, err
		}

		rawEnvByOperator[opId] = *rawEnv
	}

	txn, err := h.Database.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer database.TxnRollbackIgnoreErr(ctx, txn)

	envByOperator, err := exec_env.CreateMissingAndSyncExistingEnvs(
		ctx,
		h.ExecutionEnvironmentReader,
		h.ExecutionEnvironmentWriter,
		rawEnvByOperator,
		txn,
	)
	if err != nil {
		return nil, err
	}

	if err := txn.Commit(ctx); err != nil {
		return nil, err
	}

	return envByOperator, nil
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
