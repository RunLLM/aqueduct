/**
 * Enum representing the various loading statuses that occur when making a call to the REST API.
 */
export enum LoadingStatusEnum {
  Initial = 'initial',
  Loading = 'loading',
  Failed = 'failed',
  Succeeded = 'succeeded',
}

/**
 * Type representing a request's loading status and an error string.
 */
export type LoadingStatus = {
  loading: LoadingStatusEnum;
  err: string;
};

/**
 * Convenience function used to determine if a request is in the initial loading state.
 * @param status The loading status of the request.
 * @returns true if the request is in the initial loading state, false otherwise.
 */
export function isInitial(status: LoadingStatus): boolean {
  return status.loading === LoadingStatusEnum.Initial;
}

/**
 * Convenience function used to determine if a request is in the loading state.
 * @param status The loading status of the request.
 * @returns true if request is in the loading state, false otherwise.
 */
export function isLoading(status: LoadingStatus): boolean {
  return status.loading === LoadingStatusEnum.Loading;
}

/**
 * Convenience function used to determine if a request has successfully loaded.
 * @param status The loading status of the request.
 * @returns true if the request has successfully loaded, false otherwise.
 */
export function isSucceeded(status: LoadingStatus): boolean {
  return status.loading === LoadingStatusEnum.Succeeded;
}

/**
 * Convenience function used to determine if a request has failed to load.
 * @param status The loading status of the request.
 * @returns true if the request has failed to load, false otherwise.
 */
export function isFailed(status: LoadingStatus): boolean {
  return status.loading === LoadingStatusEnum.Failed;
}

/**
 * Enum representing the different statuses that an Operator can have while being executed.
 */
export enum ExecutionStatus {
  Unknown = 'unknown',
  Succeeded = 'succeeded',
  Failed = 'failed',
  Pending = 'pending',
  Canceled = 'canceled',
  Registered = 'registered',
  Running = 'running',
}

/**
 * Type representing when each state transition occurs for a given operator.
 */
export type ExecutionTimestamps = {
  registered_at?: string;
  pending_at?: string;
  running_at?: string;
  finished_at?: string;
};

/**
 * Type representing a given operator's lifecyle during execution.
 */
export type ExecState = {
  /**
   * Current execution status of the operator.
   */
  status: ExecutionStatus;
  /**
   * Whether or not the operator has failed to execute.
   */
  failure_type?: FailureType;
  /**
   * Stack trace for execution error and useful tip for user to fix the execution error.
   */
  error?: Error;
  /**
   * Logs that may appear in stdin or stdout
   */
  user_logs?: Logs;
  /**
   * Times when each state transition occured for the operator.
   */
  timestamps?: ExecutionTimestamps;
};

/**
 * Enum representing the different levels of error severity that an operator can have while executing.
 */
export enum FailureType {
  Succeess = 0,
  System = 1,
  UserFatal = 2,
  UserNonFatal = 3,
}

/**
 * Enum represnting whether or not a Check operator has passed.
 */
export enum CheckStatus {
  Succeeded = 'True',
  Failed = 'False',
}

export default ExecutionStatus;
export const TransitionLengthInMs = 200;

export const WidthTransition = `width ${TransitionLengthInMs}ms ease-in-out`;

/**
 * User generated logs to show for a given operator.
 */
export type Logs = {
  stdout?: string;
  stderr?: string;
};

/**
 * Type representing an operator execution error. Contains a stack trace as well as a useful tip to show the user to resolve their error.
 */
export type Error = {
  context?: string;
  tip?: string;
};

/**
 * Link used to file an issue against the Aqueduct open source repo.
 */
export const GithubIssueLink = `https://github.com/aqueducthq/aqueduct/issues/new?assignees=&labels=bug&template=bug_report.md&title=%5BBUG%5D`;
