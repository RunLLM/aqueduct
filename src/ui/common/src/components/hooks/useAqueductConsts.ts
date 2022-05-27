import getConfig from 'next/config';

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

    if (process.env && Object.keys(process.env).length > 0) {
        // This is being run on the server side.
        apiAddress = process.env.GATEWAY_ADDRESS;
        httpProtocol = process.env.NEXT_PUBLIC_PROTOCOL;
        awsAccountId = process.env.AWS_ACCOUNT_ID;
    } else {
        // This is being run on the client side.
        const { publicRuntimeConfig } = getConfig();
        apiAddress = publicRuntimeConfig.apiAddress;
        httpProtocol = publicRuntimeConfig.httpProtocol;
        awsAccountId = publicRuntimeConfig.awsAccountId;
    }

    return {
        apiAddress,
        httpProtocol,
        awsAccountId,
    };
};
