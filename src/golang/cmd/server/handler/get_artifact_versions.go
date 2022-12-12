package handler

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/lib/collections/artifact_result"
	"github.com/aqueducthq/aqueduct/lib/collections/operator/connector"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

// Route: /artifact/versions
// Method: GET
// Params: None
// Request:
//	Headers:
//		`api-key`: user's API Key
// Response:
//	Body:
//		serialized `getArtifactVersionsResponse`

type getArtifactVersionsResponse struct {
	LatestVersions     map[uuid.UUID]artifactVersions `json:"latest_versions"`
	HistoricalVersions map[uuid.UUID]artifactVersions `json:"historical_versions"`
}

type artifactVersions struct {
	WorkflowName        string                        `json:"workflow_name"`
	ArtifactName        string                        `json:"artifact_name"`
	WorkflowId          uuid.UUID                     `json:"workflow_id"`
	WorkflowDagResultId uuid.UUID                     `json:"workflow_dag_result_id"`
	ArtifactId          uuid.UUID                     `json:"artifact_id"`
	LoadSpecs           []connector.Load              `json:"load_specs"`
	Versions            map[uuid.UUID]artifactVersion `json:"versions"`
}

type artifactVersion struct {
	Timestamp int64                     `json:"timestamp"`
	Status    shared.ExecutionStatus    `json:"status"`
	Error     string                    `json:"error"`
	Metadata  *artifact_result.Metadata `json:"metadata"`
	Checks    []CheckResult             `json:"checks"`
}

type CheckResult struct {
	Name     string                 `json:"name"`
	Status   shared.ExecutionStatus `json:"status"`
	Metadata *shared.ExecutionState `json:"metadata"`
}

type GetArtifactVersionsHandler struct {
	GetHandler

	Database database.Database

	ArtifactRepo       repos.Artifact
	ArtifactResultRepo repos.ArtifactResult
	DAGRepo            repos.DAG
	OperatorRepo       repos.Operator
	OperatorResultRepo repos.OperatorResult
}

func (*GetArtifactVersionsHandler) Name() string {
	return "GetArtifactVersions"
}

func (*GetArtifactVersionsHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	return aqContext, http.StatusOK, nil
}

func (h *GetArtifactVersionsHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*aq_context.AqContext)

	emptyResponse := getArtifactVersionsResponse{}

	latestVersions := make(map[uuid.UUID]artifactVersions)
	historicalVersions := make(map[uuid.UUID]artifactVersions)

	loadSpecs, err := h.OperatorRepo.GetLoadOPSpecsByOrg(
		ctx,
		args.OrgID,
		h.Database,
	)
	if err != nil {
		return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Unable to get artifact versions.")
	}

	loadOperatorIDs := make([]uuid.UUID, 0, len(loadSpecs))
	for _, loadOperator := range loadSpecs {
		loadOperatorIDs = append(loadOperatorIDs, loadOperator.OperatorID)
	}

	latestDagIDs, err := h.DAGRepo.GetLatestIDsByOrg(
		ctx,
		args.OrgID,
		h.Database,
	)
	if err != nil {
		return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Unable to get artifact versions.")
	}

	// Handle the no data case so it doesn't throw an error in GetArtifactIdsFromWorkflowDagIdsAndDownstreamOperatorIds
	if len(latestDagIDs) == 0 || len(loadOperatorIDs) == 0 {
		return getArtifactVersionsResponse{
			LatestVersions:     latestVersions,
			HistoricalVersions: historicalVersions,
		}, http.StatusOK, nil
	}

	// We separate artifacts that belong to the latest workflow DAG from historical artifacts because we
	// want to show them separately on the UI.
	artifactIDs, err := h.ArtifactRepo.GetIDsByDAGAndDownstreamOPBatch(
		ctx,
		latestDagIDs,
		loadOperatorIDs,
		h.Database,
	)
	if err != nil {
		return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Unable to get artifact versions.")
	}

	latestArtifactIDs := make(map[uuid.UUID]bool, len(artifactIDs))
	for _, artifactID := range artifactIDs {
		latestArtifactIDs[artifactID] = true
	}

	allArtifactIds := make([]uuid.UUID, 0, len(loadOperatorIDs))

	for _, loadOperator := range loadSpecs {
		if _, ok := latestArtifactIDs[loadOperator.ArtifactID]; ok {
			// If we reach here, it means this load operator corresponds to an artifact that belongs
			// to the latest workflow dag of a workflow.
			if _, ok := latestVersions[loadOperator.ArtifactID]; !ok {
				// If we reach here, it means `latestVersions` doesn't have the artifact id entry yet,
				// so we initialize the `artifactVersions` object with the workflow name and an empty
				// `Versions` map.
				allArtifactIds = append(allArtifactIds, loadOperator.ArtifactID)
				latestVersions[loadOperator.ArtifactID] = artifactVersions{
					WorkflowName: loadOperator.WorkflowName,
					WorkflowId:   loadOperator.WorkflowID,
					ArtifactName: loadOperator.ArtifactName,
					ArtifactId:   loadOperator.ArtifactID,
					Versions:     make(map[uuid.UUID]artifactVersion),
				}
			}

			artifactVersionsObject := latestVersions[loadOperator.ArtifactID]
			artifactVersionsObject.LoadSpecs = append(artifactVersionsObject.LoadSpecs, *loadOperator.Spec.Load())

			latestVersions[loadOperator.ArtifactID] = artifactVersionsObject
		} else {
			if _, ok := historicalVersions[loadOperator.ArtifactID]; !ok {
				allArtifactIds = append(allArtifactIds, loadOperator.ArtifactID)
				historicalVersions[loadOperator.ArtifactID] = artifactVersions{
					WorkflowName: loadOperator.WorkflowName,
					Versions:     make(map[uuid.UUID]artifactVersion),
				}
			}

			artifactVersionsObject := historicalVersions[loadOperator.ArtifactID]
			artifactVersionsObject.LoadSpecs = append(artifactVersionsObject.LoadSpecs, *loadOperator.Spec.Load())

			historicalVersions[loadOperator.ArtifactID] = artifactVersionsObject
		}
	}

	artifactResultStatuses, err := h.ArtifactResultRepo.GetStatusByArtifactBatch(ctx, allArtifactIds, h.Database)
	if err != nil {
		return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Unable to get artifact versions.")
	}

	// We track failed artifact versions and later on issue another query to fetch the
	// corresponding operator's error message.
	failedArtifactIdsMap := make(map[uuid.UUID]bool, len(allArtifactIds))
	failedDAGResultIDs := make([]uuid.UUID, 0, len(artifactResultStatuses))

	// Create artifact versions and add timestamp and status metadata to it.
	// We leave the validation test result slice empty for now.
	for _, artifactResultStatus := range artifactResultStatuses {
		var artifactVersionObject artifactVersion
		if artifactResultStatus.Metadata.IsNull {
			artifactVersionObject = artifactVersion{
				Timestamp: artifactResultStatus.Timestamp.Unix(),
				Status:    artifactResultStatus.Status,
				Checks:    nil,
			}
		} else {
			artifactVersionObject = artifactVersion{
				Timestamp: artifactResultStatus.Timestamp.Unix(),
				Status:    artifactResultStatus.Status,
				Metadata:  &artifactResultStatus.Metadata.Metadata,
				Checks:    nil,
			}
		}

		if _, ok := latestVersions[artifactResultStatus.ArtifactID]; ok {
			latestVersions[artifactResultStatus.ArtifactID].Versions[artifactResultStatus.DAGResultID] = artifactVersionObject
		} else {
			historicalVersions[artifactResultStatus.ArtifactID].Versions[artifactResultStatus.DAGResultID] = artifactVersionObject
		}

		if artifactResultStatus.Status == shared.FailedExecutionStatus {
			failedArtifactIdsMap[artifactResultStatus.ArtifactID] = true
			failedDAGResultIDs = append(failedDAGResultIDs, artifactResultStatus.DAGResultID)
		}
	}

	failedArtifactIDs := make([]uuid.UUID, 0, len(failedArtifactIdsMap))
	for failedArtifactId := range failedArtifactIdsMap {
		failedArtifactIDs = append(failedArtifactIDs, failedArtifactId)
	}

	checkStatuses, err := h.OperatorResultRepo.GetCheckStatusByArtifactBatch(
		ctx,
		allArtifactIds,
		h.Database,
	)
	if err != nil {
		return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Unable to get artifact versions.")
	}

	// We now fill in the check test result. We can join the validation test result with the correct
	// artifact version by the workflow dag result id.
	for _, checkStatus := range checkStatuses {
		checkResultObject := CheckResult{
			Name:     checkStatus.OperatorName.String,
			Status:   checkStatus.Metadata.Status,
			Metadata: checkStatus.Metadata,
		}

		if _, ok := latestVersions[checkStatus.ArtifactID]; ok {
			artifactVersionObject := latestVersions[checkStatus.ArtifactID].Versions[checkStatus.DAGResultID]
			artifactVersionObject.Checks = append(artifactVersionObject.Checks, checkResultObject)
			latestVersions[checkStatus.ArtifactID].Versions[checkStatus.DAGResultID] = artifactVersionObject
		} else {
			artifactVersionObject := historicalVersions[checkStatus.ArtifactID].Versions[checkStatus.DAGResultID]
			artifactVersionObject.Checks = append(artifactVersionObject.Checks, checkResultObject)
			historicalVersions[checkStatus.ArtifactID].Versions[checkStatus.DAGResultID] = artifactVersionObject
		}
	}

	// Issue query to fetch error message only when there is at least one failed artifact version.
	if len(failedArtifactIDs) > 0 {
		failedStatuses, err := h.OperatorResultRepo.GetStatusByDAGResultAndArtifactBatch(
			ctx,
			failedDAGResultIDs,
			failedArtifactIDs,
			h.Database,
		)
		if err != nil {
			return emptyResponse, http.StatusInternalServerError, errors.Wrap(err, "Unable to get artifact versions.")
		}

		for _, failedStatus := range failedStatuses {
			if _, ok := latestVersions[failedStatus.ArtifactID]; ok {
				artifactVersionObject := latestVersions[failedStatus.ArtifactID].Versions[failedStatus.DAGResultID]
				if failedStatus.Metadata != nil && failedStatus.Metadata.Error != nil {
					artifactVersionObject.Error = failedStatus.Metadata.Error.Context
				}

				latestVersions[failedStatus.ArtifactID].Versions[failedStatus.DAGResultID] = artifactVersionObject
			} else {
				artifactVersionObject := historicalVersions[failedStatus.ArtifactID].Versions[failedStatus.DAGResultID]
				if failedStatus.Metadata != nil && failedStatus.Metadata.Error != nil {
					artifactVersionObject.Error = failedStatus.Metadata.Error.Context
				}

				historicalVersions[failedStatus.ArtifactID].Versions[failedStatus.DAGResultID] = artifactVersionObject
			}
		}
	}

	return getArtifactVersionsResponse{
		LatestVersions:     latestVersions,
		HistoricalVersions: historicalVersions,
	}, http.StatusOK, nil
}
