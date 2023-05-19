import { APIKeyParameter } from '../parameters/Header';

export type EnvironmentGetRequest = APIKeyParameter;

export type EnvironmentGetResponse = {
  inK8sCluster: boolean;
  version: string;
};

export const environmentGetQuery = (req: EnvironmentGetRequest) => ({
  url: 'environment',
  headers: { 'api-key': req.apiKey },
});
