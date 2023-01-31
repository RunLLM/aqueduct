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
}

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
    default:
      executionStatus = ExecutionStatus.Unknown;
      break;
  }

  return executionStatus;
}

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
