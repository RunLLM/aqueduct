// This file should map exactly to
// `src/golang/lib/response/node.go`
import { ArtifactType } from '../../utils/artifacts';
import { OperatorSpec } from '../../utils/operators';

export type ArtifactResponse = {
  id: string;
  dag_id: string;
  name: string;
  description: string;
  type: ArtifactType;
  inputs: string[];
  outputs: string[];
};

export type OperatorResponse = {
  id: string;
  dag_id: string;
  name: string;
  description: string;
  spec?: OperatorSpec;
  inputs: string[];
  outputs: string[];
};

export type NodesResponse = {
  operators: OperatorResponse[];
  artifacts: ArtifactResponse[];
};
