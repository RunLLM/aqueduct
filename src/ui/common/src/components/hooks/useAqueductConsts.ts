export type ClusterEnvironment = 'dev' | 'staging' | 'prod';

export type AqueductConsts = {
  apiAddress: string;
  httpProtocol: string;
  awsAccountId: string;
};

export const useAqueductConsts = (): AqueductConsts => {
  let apiAddress;
  let httpProtocol;
  let awsAccountId;

  apiAddress = process.env.GATEWAY_ADDRESS;
  httpProtocol = process.env.NEXT_PUBLIC_PROTOCOL;
  awsAccountId = process.env.AWS_ACCOUNT_ID;

  return {
    apiAddress,
    httpProtocol,
    awsAccountId,
  };
};
