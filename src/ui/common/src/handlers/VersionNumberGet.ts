import { APIKeyRequest } from './requests/ApiKey';

export type VersionNumberGetRequest = APIKeyRequest;

export type VersionNumberGetResponse = { version: string };

export const versionNumberGetQuery = (req: VersionNumberGetRequest) => ({
  url: `version`,
  headers: { 'api-key': req.apiKey },
});
