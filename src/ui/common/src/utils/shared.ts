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

export enum ExecutionStatus {
  Succeeded = 'succeeded',
  Failed = 'failed',
  Pending = 'pending',
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
