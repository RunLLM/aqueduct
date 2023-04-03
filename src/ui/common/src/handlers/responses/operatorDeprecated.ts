// This file should mirror src/golang/workflow/operator/response.go
import { OperatorSpec } from '../../utils/operators';
import { ExecState } from '../../utils/shared';

export type OperatorResponse = {
  id: string;
  name: string;
  description: string;
  spec?: OperatorSpec;
  inputs: string[];
  outputs: string[];
};

export type OperatorResultStatusResponse = {
  id: string;
  exec_state?: ExecState;
};

export type OperatorResultResponse = OperatorResponse & {
  result?: OperatorResultStatusResponse;
};
