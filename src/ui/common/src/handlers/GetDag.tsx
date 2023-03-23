import { APIKeyRequest } from './requests/ApiKey';
import { DagResponse } from './responses/dag';

export type GetDagRequest = APIKeyRequest & {
  workflowId: string;
  dagId: string;
};

export type GetDagResponse = DagResponse;

export const getDagQuery = (req: GetDagRequest) => ({
  url: `workflow/${req.workflowId}/dag/${req.dagId}`,
  headers: { 'api-key': req.apiKey },
});
