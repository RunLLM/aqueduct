import { APIKeyParameter } from '../parameters/Header';

export type EnvironmentGetRequest = APIKeyParameter;

export type EnvironeGetResponse = {
    inK8sCluster: boolean;
};

// TODO: Move this endpoint to the v2 API
export const dagOperatorsGetQuery = (req: EnvironmentGetRequest) => ({
  //url: `workflows/${req.workflowId}/dag/${req.dagId}/nodes/operators`,
  url: 'environment',
  headers: { 'api-key': req.apiKey },
});
