export type AqueductConsts = {
  apiAddress: string;
  httpProtocol: string;
};

export const useAqueductConsts = (): AqueductConsts => {
  let apiAddress;
  let httpProtocol;

  apiAddress = process.env.SERVER_ADDRESS;
  httpProtocol = process.env.NEXT_PUBLIC_PROTOCOL;

  return {
    apiAddress,
    httpProtocol,
  };
};
