// This file should map exactly to
// src/golang/cmd/server/handler/v2/workflow_post.go

import { APIKeyParameter } from '../parameters/Header';
import { WorkflowIdParameter } from '../parameters/Path';

export type WorkflowTriggerPostRequest = APIKeyParameter &
  WorkflowIdParameter & {
    serializedParams: string;
  };

export type WorkflowTriggerPostResponse = Record<string, never>;

export const workflowTriggerPostQuery = (req: WorkflowTriggerPostRequest) => {
  const parameters = new FormData();
  parameters.append('parameters', req.serializedParams);
  return {
    url: `workflow/${req.workflowId}/edit`,
    method: 'POST',
    // avoid built-in content-type override
    // ref: https://github.com/reduxjs/redux-toolkit/issues/2287
    prepareHeaders: (headers) => {
      headers.set('api-key', req.apiKey);
      headers.set('Content-Type', 'multipart/form-data');
      return headers;
    },
    body: parameters,
  };
};
