import { emptySplitApi as api } from "../../rtk_placeholder";
const injectedRtkApi = api.injectEndpoints({
  endpoints: (build) => ({
    workflowGet: build.query<WorkflowGetApiResponse, WorkflowGetApiArg>({
      query: (queryArg) => ({
        url: `/workflow/${queryArg.workflowId}`,
        headers: { "api-key": queryArg["api-key"] },
      }),
    }),
  }),
  overrideExisting: false,
});
export { injectedRtkApi as aqueduct };
export type WorkflowGetApiResponse =
  /** status 200 The metadata of the given workflow. */ Workflow;
export type WorkflowGetApiArg = {
  /** the ID of workflow object */
  workflowId: string;
  /** the user's API Key */
  "api-key": string;
};
export type Schedule = {
  trigger: "manual" | "periodic" | "airflow" | "cascade";
  cron_schedule?: string;
  disable_manual_trigger?: boolean;
  paused?: boolean;
  source_id?: string;
};
export type RetentionPolicy = {
  k_latest_runs: number;
};
export type NotificationSettings = {
  integration_id: string;
  notification_level: "success" | "warning" | "error" | "info" | "neutral";
}[];
export type Workflow = {
  id: string;
  name: string;
  description?: string;
  schedule?: Schedule;
  created_at?: string;
  retention_policy?: RetentionPolicy;
  notification_settings?: NotificationSettings;
};
export type GeneralError = {
  error?: string;
};
export const { useWorkflowGetQuery } = injectedRtkApi;
