import { APIKeyParameter } from '../parameters/Header';

export type EnvironmentGetRequest = APIKeyParameter;

export type EnvironmentGetResponse = {
    inK8sCluster: boolean;
};

// TODO: Move this endpoint to the v2 API
export const environmentGetQuery = (req: EnvironmentGetRequest) => ({
  //url: `workflows/${req.workflowId}/dag/${req.dagId}/nodes/operators`,
  url: 'environment',
  headers: { 'api-key': req.apiKey },
});
