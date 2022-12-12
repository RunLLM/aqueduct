package handler

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/lib/collections/operator/connector"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/storage"
	"github.com/aqueducthq/aqueduct/lib/workflow/artifact"
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
	WorkflowID          uuid.UUID                     `json:"workflow_id"`
	WorkflowDagResultID uuid.UUID                     `json:"workflow_dag_result_id"`
	ArtifactID          uuid.UUID                     `json:"artifact_id"`
	LoadSpecs           []connector.Load              `json:"load_specs"`
	Versions            map[uuid.UUID]artifactVersion `json:"versions"`
}

type artifactVersion struct {
	Timestamp int64                     `json:"timestamp"`
	Status    shared.ExecutionStatus    `json:"status"`
	Error     string                    `json:"error"`
	Checks    []CheckResult             `json:"checks"`
	Metrics   []artifact.ResultResponse `json:"metrics"`
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

func (h *GetArtifactVersionsHandler) initializeLatestAndHistoricalVersions(
	ctx context.Context,
	orgID string,
) (
	map[uuid.UUID]artifactVersions, // latestVersions
	map[uuid.UUID]artifactVersions, // historicalVersions
	[]uuid.UUID, // all artifact IDs
	error,
) {
	latestVersions := make(map[uuid.UUID]artifactVersions)
	historicalVersions := make(map[uuid.UUID]artifactVersions)

	loadSpecs, err := h.OperatorRepo.GetLoadOPSpecsByOrg(
		ctx,
		orgID,
		h.Database,
	)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "Unable to get artifact versions.")
	}

	loadOperatorIDs := make([]uuid.UUID, 0, len(loadSpecs))
	for _, loadOperator := range loadSpecs {
		loadOperatorIDs = append(loadOperatorIDs, loadOperator.OperatorID)
	}

	latestDagIDs, err := h.DAGRepo.GetLatestIDsByOrg(
		ctx,
		orgID,
		h.Database,
	)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "Unable to get artifact versions.")
	}

	// Handle the no data case so it doesn't throw an error in GetArtifactIdsFromWorkflowDagIdsAndDownstreamOperatorIds
	if len(latestDagIDs) == 0 || len(loadOperatorIDs) == 0 {
		return nil, nil, nil, nil
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
		return nil, nil, nil, errors.Wrap(err, "Unable to get artifact versions.")
	}

	latestArtifactIDs := make(map[uuid.UUID]bool, len(artifactIDs))
	for _, artifactID := range artifactIDs {
		latestArtifactIDs[artifactID] = true
	}

	allArtifactIDs := make([]uuid.UUID, 0, len(loadOperatorIDs))

	for _, loadOperator := range loadSpecs {
		if _, ok := latestArtifactIDs[loadOperator.ArtifactID]; ok {
			// If we reach here, it means this load operator corresponds to an artifact that belongs
			// to the latest workflow dag of a workflow.
			if _, ok := latestVersions[loadOperator.ArtifactID]; !ok {
				// If we reach here, it means `latestVersions` doesn't have the artifact id entry yet,
				// so we initialize the `artifactVersions` object with the workflow name and an empty
				// `Versions` map.
				allArtifactIDs = append(allArtifactIDs, loadOperator.ArtifactID)
				latestVersions[loadOperator.ArtifactID] = artifactVersions{
					WorkflowName: loadOperator.WorkflowName,
					WorkflowID:   loadOperator.WorkflowID,
					ArtifactName: loadOperator.ArtifactName,
					ArtifactID:   loadOperator.ArtifactID,
					Versions:     make(map[uuid.UUID]artifactVersion),
				}
			}

			artifactVersionsObject := latestVersions[loadOperator.ArtifactID]
			artifactVersionsObject.LoadSpecs = append(artifactVersionsObject.LoadSpecs, *loadOperator.Spec.Load())

			latestVersions[loadOperator.ArtifactID] = artifactVersionsObject
		} else {
			if _, ok := historicalVersions[loadOperator.ArtifactID]; !ok {
				allArtifactIDs = append(allArtifactIDs, loadOperator.ArtifactID)
				historicalVersions[loadOperator.ArtifactID] = artifactVersions{
					WorkflowName: loadOperator.WorkflowName,
					WorkflowID:   loadOperator.WorkflowID,
					ArtifactName: loadOperator.ArtifactName,
					ArtifactID:   loadOperator.ArtifactID,
					Versions:     make(map[uuid.UUID]artifactVersion),
				}
			}

			artifactVersionsObject := historicalVersions[loadOperator.ArtifactID]
			artifactVersionsObject.LoadSpecs = append(artifactVersionsObject.LoadSpecs, *loadOperator.Spec.Load())

			historicalVersions[loadOperator.ArtifactID] = artifactVersionsObject
		}
	}

	return latestVersions, historicalVersions, allArtifactIDs, nil
}

func (h *GetArtifactVersionsHandler) updateVersionsWithArtifactResultStatuses(
	ctx context.Context,
	latestVersions map[uuid.UUID]artifactVersions,
	historicalVersions map[uuid.UUID]artifactVersions,
	allArtifactIDs []uuid.UUID,
) (
	[]uuid.UUID, // failedArtifactIDs
	[]uuid.UUID, // failedDAGResultIDs
	error,
) {
	artifactResultStatuses, err := h.ArtifactResultRepo.GetStatusByArtifactBatch(ctx, allArtifactIDs, h.Database)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Unable to get artifact versions.")
	}

	failedArtifactIDsMap := make(map[uuid.UUID]bool, len(allArtifactIDs))
	failedDAGResultIDs := make([]uuid.UUID, 0, len(artifactResultStatuses))

	// Create artifact versions and add timestamp and status metadata to it.
	// We leave the validation test result slice empty for now.
	for _, artifactResultStatus := range artifactResultStatuses {
		if _, ok := latestVersions[artifactResultStatus.ArtifactID]; ok {
			latestVersions[artifactResultStatus.ArtifactID].Versions[artifactResultStatus.DAGResultID] = artifactVersion{
				Timestamp: artifactResultStatus.Timestamp.Unix(),
				Status:    artifactResultStatus.Status,
				Checks:    nil,
			}
		} else {
			historicalVersions[artifactResultStatus.ArtifactID].Versions[artifactResultStatus.DAGResultID] = artifactVersion{
				Timestamp: artifactResultStatus.Timestamp.Unix(),
				Status:    artifactResultStatus.Status,
				Checks:    nil,
			}
		}

		if artifactResultStatus.Status == shared.FailedExecutionStatus {
			failedArtifactIDsMap[artifactResultStatus.ArtifactID] = true
			failedDAGResultIDs = append(failedDAGResultIDs, artifactResultStatus.DAGResultID)
		}
	}

	failedArtifactIDs := make([]uuid.UUID, 0, len(failedArtifactIDsMap))
	for failedArtifactId := range failedArtifactIDsMap {
		failedArtifactIDs = append(failedArtifactIDs, failedArtifactId)
	}

	return failedArtifactIDs, failedDAGResultIDs, nil
}

func (h *GetArtifactVersionsHandler) updateVersionsWithChecksAndMetrics(
	ctx context.Context,
	latestVersions map[uuid.UUID]artifactVersions,
	historicalVersions map[uuid.UUID]artifactVersions,
	allArtifactIDs []uuid.UUID,
) error {
	checkStatuses, err := h.OperatorResultRepo.GetCheckStatusByArtifactBatch(
		ctx,
		allArtifactIDs,
		h.Database,
	)
	if err != nil {
		return errors.Wrap(err, "Unable to get artifact versions.")
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

	metricsByUpstreamArtifactID, err := h.ArtifactRepo.GetMetricsByUpstreamArtifactBatch(ctx, allArtifactIDs, h.Database)
	if err != nil {
		return err
	}

	metricsIDs := []uuid.UUID{}
	metricsMap := map[uuid.UUID]models.Artifact{}
	for _, metrics := range metricsByUpstreamArtifactID {
		for _, metric := range metrics {
			metricsIDs = append(metricsIDs, metric.ID)
			metricsMap[metric.ID] = metric
		}
	}

	metricResults, err := h.ArtifactResultRepo.GetByArtifactBatch(ctx, metricsIDs, h.Database)
	if err != nil {
		return err
	}

	metricResultsIDs := make([]uuid.UUID, 0, len(metricResults))
	metricResultsByArtfID := make(map[uuid.UUID][]models.ArtifactResult, len(metricResults))
	for _, metricResult := range metricResults {
		metricResultsIDs = append(metricResultsIDs, metricResult.ID)
		metricResultsByArtfID[metricResult.ArtifactID] = append(
			metricResultsByArtfID[metricResult.ArtifactID],
			metricResult,
		)
	}

	dagsByMetricResultID, err := h.DAGRepo.GetByArtifactResultBatch(ctx, metricResultsIDs, h.Database)
	if err != nil {
		return err
	}

	for upstreamArtifactID, metrics := range metricsByUpstreamArtifactID {
		for _, metric := range metrics {
			metricResults, ok := metricResultsByArtfID[metric.ID]
			if !ok {
				continue
			}

			for _, metricResult := range metricResults {
				var contentPtr *string = nil
				dag, ok := dagsByMetricResultID[metricResult.ID]
				if ok {
					storageObj := storage.NewStorage(&dag.StorageConfig)
					if metric.Type.IsCompact() {
						path := metricResult.ContentPath
						contentBytes, err := storageObj.Get(ctx, path)
						if err == nil {
							content := string(contentBytes)
							contentPtr = &content
						} else if err != storage.ErrObjectDoesNotExist {
							return errors.Wrap(err, "Unable to get artifact content from storage")
						}
					}
				}

				metricResp := artifact.NewResultResponseFromDbObjects(
					&metric,
					&metricResult,
					contentPtr,
					uuid.UUID{}, // from, we ignore this field for now
					nil,         // to, we ignore this field for now
				)

				if _, ok := latestVersions[upstreamArtifactID]; ok {
					artifactVersionObject := latestVersions[upstreamArtifactID].Versions[metricResult.DAGResultID]
					artifactVersionObject.Metrics = append(artifactVersionObject.Metrics, *metricResp)
					latestVersions[upstreamArtifactID].Versions[metricResult.DAGResultID] = artifactVersionObject
				} else {
					artifactVersionObject := historicalVersions[upstreamArtifactID].Versions[metricResult.DAGResultID]
					artifactVersionObject.Metrics = append(artifactVersionObject.Metrics, *metricResp)
					historicalVersions[upstreamArtifactID].Versions[metricResult.DAGResultID] = artifactVersionObject
				}
			}
		}
	}
	return nil
}

func (h *GetArtifactVersionsHandler) updateVersionsWithErrorMessages(
	ctx context.Context,
	latestVersions map[uuid.UUID]artifactVersions,
	historicalVersions map[uuid.UUID]artifactVersions,
	failedDAGResultIDs []uuid.UUID,
	failedArtifactIDs []uuid.UUID,
) error {
	failedStatuses, err := h.OperatorResultRepo.GetStatusByDAGResultAndArtifactBatch(
		ctx,
		failedDAGResultIDs,
		failedArtifactIDs,
		h.Database,
	)
	if err != nil {
		return errors.Wrap(err, "Unable to get artifact versions.")
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

	return nil
}

func (h *GetArtifactVersionsHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*aq_context.AqContext)

	emptyResponse := getArtifactVersionsResponse{}

	latestVersions, historicalVersions, allArtifactIDs, err := h.initializeLatestAndHistoricalVersions(
		ctx, args.OrgID,
	)
	if err != nil {
		return emptyResponse, http.StatusInternalServerError, err
	}

	if len(latestVersions) == 0 && len(historicalVersions) == 0 {
		return getArtifactVersionsResponse{
			LatestVersions:     latestVersions,
			HistoricalVersions: historicalVersions,
		}, http.StatusOK, nil
	}

	// We track failed artifact versions and later on issue another query to fetch the
	// corresponding operator's error message.
	failedArtifactIDs, failedDAGResultIDs, err := h.updateVersionsWithArtifactResultStatuses(
		ctx, latestVersions, historicalVersions, allArtifactIDs,
	)
	if err != nil {
		return emptyResponse, http.StatusInternalServerError, err
	}

	err = h.updateVersionsWithChecksAndMetrics(ctx, latestVersions, historicalVersions, allArtifactIDs)
	if err != nil {
		return emptyResponse, http.StatusInternalServerError, err
	}

	// Issue query to fetch error message only when there is at least one failed artifact version.
	if len(failedArtifactIDs) > 0 {
		err = h.updateVersionsWithErrorMessages(ctx, latestVersions, historicalVersions, failedDAGResultIDs, failedArtifactIDs)
		if err != nil {
			return emptyResponse, http.StatusInternalServerError, err
		}
	}

	return getArtifactVersionsResponse{
		LatestVersions:     latestVersions,
		HistoricalVersions: historicalVersions,
	}, http.StatusOK, nil
}
