export const ContentSidebarOffsetInPx = 100;

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

export function isLoading(status: LoadingStatus) {
  return (
    status.loading === LoadingStatusEnum.Initial ||
    status.loading === LoadingStatusEnum.Loading
  );
}

export function isSucceeded(status: LoadingStatus) {
  return status.loading === LoadingStatusEnum.Succeeded;
}

export function isFailed(status: LoadingStatus) {
  return status.loading === LoadingStatusEnum.Failed;
}

export enum ExecutionStatus {
  Unknown = 'unknown',
  Succeeded = 'succeeded',
  Failed = 'failed',
  Pending = 'pending',
}

export enum FailureType {
  System = 1,
  User = 2,
}

export enum CheckStatus {
  Succeeded = 'True',
  Failed = 'False',
}

export default ExecutionStatus;
export const TransitionLengthInMs = 200;

export const WidthTransition = `width ${TransitionLengthInMs}ms ease-in-out`;
export const HeightTransition = `height ${TransitionLengthInMs}ms ease-in-out`;
export const AllTransition = `all ${TransitionLengthInMs}ms ease-in-out`;

export type Logs = {
  stdout?: string;
  stderr?: string;
};

export type Error = {
  context?: string;
  tip?: string;
};
