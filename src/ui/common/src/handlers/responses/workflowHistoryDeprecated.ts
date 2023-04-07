import { ExecState } from 'src/utils/shared';

// This file should mirror src/golang/cmd/server/handler/get_workflow_history.go
export type WorkflowVersionResponse = {
  versionId: string;
  created_at: number;
  exec_state: ExecState;
};

export type WorkflowHistoryResponse = {
  id: string;
  name: string;
  versions: WorkflowVersionResponse[];
};
