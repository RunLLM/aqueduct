import { APIKeyRequest } from './requests/ApiKey';
import { DagResultResponse } from './responses/dag';

export type GetDagResultRequest = APIKeyRequest & {
  workflowId: string;
  dagResultId: string;
};

export type GetDagResultResponse = DagResultResponse;

export const getDagResultQuery = (req: GetDagResultRequest) => ({
  url: `workflow/${req.workflowId}/result/${req.dagResultId}`,
  headers: { 'api-key': req.apiKey },
});
