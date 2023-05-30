import { ArtifactResultResponse } from '../handlers/responses/node';
import { TableRow } from './data';

export enum AWSCredentialType {
  AccessKey = 'access_key',
  ConfigFilePath = 'config_file_path',
  ConfigFileContent = 'config_file_content',
}

export enum LoadingStatusEnum {
  Initial = 'initial',
  Loading = 'loading',
  Failed = 'failed',
  Succeeded = 'succeeded',
}

export type LoadingStatus = {
  loading: LoadingStatusEnum;
  err: string;
};

export function isInitial(status: LoadingStatus): boolean {
  return status.loading === LoadingStatusEnum.Initial;
}

export function isLoading(status: LoadingStatus): boolean {
  return status.loading === LoadingStatusEnum.Loading;
}

export function isSucceeded(status: LoadingStatus): boolean {
  return status.loading === LoadingStatusEnum.Succeeded;
}

export function isFailed(status: LoadingStatus): boolean {
  return status.loading === LoadingStatusEnum.Failed;
}

export enum ExecutionStatus {
  Unknown = 'unknown',
  Succeeded = 'succeeded',
  Failed = 'failed',
  Pending = 'pending',
  Canceled = 'canceled',
  Registered = 'registered',
  Running = 'running',
  // Checks can have a warning status.
  Warning = 'warning',
}

export const getArtifactResultTableRow = (
  artifactResult: ArtifactResultResponse
): TableRow => {
  const all_times = [
    artifactResult.exec_state?.timestamps?.finished_at,
    artifactResult.exec_state?.timestamps?.pending_at,
    artifactResult.exec_state?.timestamps?.registered_at,
    artifactResult.exec_state?.timestamps?.running_at,
  ];

  const times = all_times
    .filter((x) => typeof x === 'string')
    .map((x) => new Date(x)); // Convert from string to time

  const maxTime = Math.max.apply(null, times); // Returns -Infinity if times is an empty list

  // Need to convert back to Date because math.max changes the time to Unix time (number)
  const timestamp =
    maxTime > 0 ? new Date(maxTime).toLocaleString() : 'Unknown';

  return {
    timestamp,
    status: artifactResult.exec_state?.status ?? 'Unknown',
    value: artifactResult.content_serialized,
  };
};

export const stringToExecutionStatus = (status: string): ExecutionStatus => {
  let executionStatus = ExecutionStatus.Unknown;
  switch (status) {
    case 'unknown':
      executionStatus = ExecutionStatus.Unknown;
      break;
    case 'succeeded':
      executionStatus = ExecutionStatus.Succeeded;
      break;
    case 'failed':
      executionStatus = ExecutionStatus.Failed;
      break;
    case 'pending':
      executionStatus = ExecutionStatus.Pending;
      break;
    case 'canceled':
      executionStatus = ExecutionStatus.Canceled;
      break;
    case 'registered':
      executionStatus = ExecutionStatus.Registered;
      break;
    case 'running':
      executionStatus = ExecutionStatus.Running;
      break;
    case 'warning':
      executionStatus = ExecutionStatus.Warning;
      break;
    default:
      executionStatus = ExecutionStatus.Unknown;
      break;
  }

  return executionStatus;
};

export type ExecutionTimestamps = {
  registered_at?: string;
  pending_at?: string;
  running_at?: string;
  finished_at?: string;
};

export type ExecState = {
  status: ExecutionStatus;
  failure_type?: FailureType;
  error?: Error;
  user_logs?: Logs;
  timestamps?: ExecutionTimestamps;
};

export enum FailureType {
  Succeess = 0,
  System = 1,
  UserFatal = 2,
  UserNonFatal = 3,
}

export enum CheckStatus {
  Succeeded = 'True',
  Failed = 'False',
}

export default ExecutionStatus;
export const TransitionLengthInMs = 200;

export const WidthTransition = `width ${TransitionLengthInMs}ms ease-in-out`;

export type Logs = {
  stdout?: string;
  stderr?: string;
};

export type Error = {
  context?: string;
  tip?: string;
};

export const GithubIssueLink = `https://github.com/aqueducthq/aqueduct/issues/new?assignees=&labels=bug&template=bug_report.md&title=%5BBUG%5D`;

// 0.875rem is the size of the ShowMore.
export const showMoreFontSize = '0.875rem';

// Add this additional padding if we don't have more than 1 metric to keep the rows equal size.
export const showMorePadding = `${showMoreFontSize} 0 ${showMoreFontSize} 0`;
