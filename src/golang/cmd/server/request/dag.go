package request

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/models/shared/operator/function"
	"github.com/aqueducthq/aqueduct/lib/models/utils"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/github"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

const dagKey = "dag"

type DagSummary struct {
	Dag *models.DAG

	// Extract the operator contents from the request body
	FileContentsByOperatorUUID map[uuid.UUID][]byte
}

func ParseDagSummaryFromRequest(
	r *http.Request,
	userId uuid.UUID,
	githubManager github.Manager,
	storageConfig *shared.StorageConfig,
) (*DagSummary, int, error) {
	serializedDAGBytes, err := ExtractHttpPayload(
		r.Header.Get(routes.ContentTypeHeader),
		dagKey,
		false, // not a file
		r,
	)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Serialized dag object not available")
	}

	var dag models.DAG
	err = json.Unmarshal(serializedDAGBytes, &dag)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Invalid dag specification.")
	}

	ghClient, err := githubManager.GetClient(r.Context(), userId)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	fileContents := make(map[uuid.UUID][]byte, len(dag.Operators))
	for opId, op := range dag.Operators {
		op.ExecutionEnvironmentID = utils.NullUUID{IsNull: true}
		dag.Operators[opId] = op

		program, status, err := extractOperatorContentsFromRequest(
			r,
			&op,
			ghClient,
		)
		if err != nil {
			return nil, status, err
		}

		if len(program) > 0 {
			fileContents[op.ID] = program
		}
	}

	dag.StorageConfig = *storageConfig
	dag.Metadata.UserID = userId

	return &DagSummary{
		Dag:                        &dag,
		FileContentsByOperatorUUID: fileContents,
	}, http.StatusOK, nil
}

// Extracts files and all github contents.
// For github contents, retrieve zipball for files and update string contents like sql queries.
func extractOperatorContentsFromRequest(
	r *http.Request,
	op *models.Operator,
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
		program, err := ExtractHttpPayload(
			r.Header.Get(routes.ContentTypeHeader),
			op.ID.String(), // File name should match operator ID
			true,
			r,
		)
		if err != nil {
			return nil, http.StatusBadRequest, errors.Wrap(
				err,
				fmt.Sprintf(
					"Required operator file %s doesn't exist.",
					op.ID.String(),
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
