import { apiAddress } from '../components/hooks/useAqueductConsts';
import UserProfile from './auth';

export enum NotificationStatus {
  Unread = 'unread',
  Archived = 'archived',
}

export enum NotificationLogLevel {
  Success = 'success',
  Warning = 'warning',
  Error = 'error',
  Info = 'info',
  Neutral = 'neutral',
}

export enum NotificationAssociation {
  Workflow = 'workflow',
  WorkflowDagResult = 'workflow_dag_result',
  Organization = 'organization',
}

export type NotificationWorkflowMetadata = {
  dag_result_id: string;
  id: string;
  name: string;
};

export type Notification = {
  id: string;
  content: string;
  status: NotificationStatus;
  level: NotificationLogLevel;
  association: {
    id: string;
    object: NotificationAssociation;
  };
  createdAt: number;
  workflowMetadata: NotificationWorkflowMetadata;
};

export async function listNotifications(
  user: UserProfile
): Promise<[Notification[], string]> {
  try {
    console.log("list notification: calling notification route!");
    const res = await fetch(`${apiAddress}/api/notifications`, {
      method: 'GET',
      headers: { 'api-key': user.apiKey },
    });

    const body = await res.json();
    if (!res.ok) {
      return [[], body.error];
    }

    return [body ?? [], ''];
  } catch (err) {
    return [[], err as string];
  }
}

// Returns empty string if the function succeeded, otherwise, returns the error message.
export async function archiveNotification(
  user: UserProfile,
  id: string
): Promise<string> {
  try {
    console.log("archive notification: calling notification route!");
    const res = await fetch(`${apiAddress}/api/notifications/${id}/archive`, {
      method: 'POST',
      headers: { 'api-key': user.apiKey },
    });

    const body = await res.json();
    if (!res.ok) {
      return body.error;
    }

    return '';
  } catch (err) {
    return err as string;
  }
}
