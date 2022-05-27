package request_parser

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aqueducthq/aqueduct/internal/server/utils"
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/github"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/function"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

type DagSummary struct {
	Dag *workflow_dag.WorkflowDag

	// Extract the operator contents from the request body
	FileContentsByOperatorUUID map[uuid.UUID][]byte
}

func ParseDagSummaryFromRequest(
	r *http.Request,
	userId uuid.UUID,
	githubManager github.Manager,
	storageConfig *shared.StorageConfig,
) (*DagSummary, int, error) {
	serializedDAGBytes, err := utils.ExtractHttpPayload(
		r.Header.Get(utils.ContentTypeHeader),
		dagKey,
		false, // not a file
		r,
	)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Serialized dag object not available")
	}

	var workflowDag workflow_dag.WorkflowDag
	err = json.Unmarshal(serializedDAGBytes, &workflowDag)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Invalid dag specification.")
	}

	ghClient, err := githubManager.GetClient(r.Context(), userId)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	fileContents := make(map[uuid.UUID][]byte, len(workflowDag.Operators))
	for _, op := range workflowDag.Operators {
		program, status, err := extractOperatorContentsFromRequest(
			r,
			&op,
			userId,
			ghClient,
		)
		if err != nil {
			return nil, status, err
		}

		if len(program) > 0 {
			fileContents[op.Id] = program
		}
	}

	workflowDag.StorageConfig = *storageConfig

	workflowDag.Metadata.UserId = userId

	return &DagSummary{
		Dag:                        &workflowDag,
		FileContentsByOperatorUUID: fileContents,
	}, http.StatusOK, nil
}

// Extracts files and all github contents.
// For github contents, retrieve zipball for files and update string contents like sql queries.
func extractOperatorContentsFromRequest(
	r *http.Request,
	op *operator.Operator,
	userId uuid.UUID,
	ghClient github.Client,
) ([]byte, int, error) {
	if op.Spec.IsExtract() {
		if github.IsExtractFromGithub(op.Spec.Extract()) {
			_, err := ghClient.PullExtract(r.Context(), op.Spec.Extract())
			if err != nil {
				return nil, http.StatusInternalServerError, err
			}
		}
		return nil, http.StatusOK, nil
	}

	if !op.Spec.HasFunction() {
		return nil, http.StatusOK, nil
	}

	fn := op.Spec.Function()

	if fn.Type == function.FileFunctionType {
		program, err := utils.ExtractHttpPayload(
			r.Header.Get(utils.ContentTypeHeader),
			op.Id.String(), // File name should match operator ID
			true,
			r,
		)
		if err != nil {
			return nil, http.StatusBadRequest, errors.Wrap(
				err,
				fmt.Sprintf(
					"Required operator file %s doesn't exist.",
					op.Id.String(),
				),
			)
		}
		return program, http.StatusOK, nil
	}

	isGhFunction, err := github.IsFunctionFromGithub(fn)
	if err == github.ErrGithubMetadataMissing {
		return nil, http.StatusBadRequest, err
	}

	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	if !isGhFunction {
		return nil, http.StatusOK, nil
	}

	_, zipball, err := ghClient.PullAndUpdateFunction(
		r.Context(),
		fn,
		true, /* alwaysPullContent */
	)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return zipball, http.StatusOK, nil
}
