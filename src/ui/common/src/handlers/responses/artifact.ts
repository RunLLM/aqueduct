// This file should mirror src/golang/workflow/artifact/response.go
import { ArtifactType } from '../../utils/artifacts';
import { ExecState } from '../../utils/shared';

export type ArtifactResponse = {
  id: string;
  name: string;
  description: string;
  type: ArtifactType;
};

export type ArtifactResultStatusResponse = {
  id: string;
  content_path: string;
  exec_state?: ExecState;
};

export type ArtifactResultResponse = ArtifactResponse & {
  result?: ArtifactResultStatusResponse;
};
