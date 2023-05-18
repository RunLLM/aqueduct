// This file should map exactly to
// src/golang/cmd/server/handler/v2/workflow_delete.go

import { SavedObjectDeletion } from '../../utils/workflows';
import { APIKeyParameter } from '../parameters/Header';
import { WorkflowIdParameter } from '../parameters/Path';

export type WorkflowDeletePostRequest = APIKeyParameter &
  WorkflowIdParameter & {
    external_delete: { [integration_id: string]: string[] };
    force: boolean;
  };

export type WorkflowDeletePostResponse = {
  [id: string]: SavedObjectDeletion;
};

export const workflowDeletePostQuery = (req: WorkflowDeletePostRequest) => ({
  url: `workflow/${req.workflowId}/delete`,
  method: 'POST',
  headers: { 'api-key': req.apiKey },
  body: {
    external_delete: req.external_delete,
    force: req.force,
  },
});
