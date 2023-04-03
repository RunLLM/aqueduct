// This file should map exactly to
// `src/golang/lib/workflow/responses`

import {
  NotificationSettings,
  RetentionPolicy,
  WorkflowSchedule,
} from '../../utils/workflows';

export type WorkflowResponse = {
  id: string;
  user_id: string;
  name: string;
  description: string;
  schedule: WorkflowSchedule;
  created_at: string;
  retention_policy: RetentionPolicy;
  notification_settings: NotificationSettings;
};
