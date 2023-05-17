// This file should map exactly to
// src/golang/cmd/server/handler/edit_workflow.go

import {
  NotificationSettings,
  RetentionPolicy,
  WorkflowSchedule,
} from '../../utils/workflows';
import { APIKeyParameter } from '../parameters/Header';
import { WorkflowIdParameter } from '../parameters/Path';

export type WorkflowEditPostRequest = APIKeyParameter &
  WorkflowIdParameter & {
    name: string;
    description: string;
    schedule: WorkflowSchedule;
    retention_policy: RetentionPolicy;
    notification_settings: NotificationSettings;
  };

export type WorkflowEditPostResponse = Record<string, never>;

export const workflowEditPostQuery = (req: WorkflowEditPostRequest) => ({
  url: `workflow/${req.workflowId}/edit`,
  method: 'POST',
  headers: { 'api-key': req.apiKey },
  body: {
    name: req.name,
    description: req.description,
    schedule: req.schedule,
    retention_policy: req.retention_policy,
    notification_settings: req.notification_settings,
  },
});
