export type AqueductConsts = {
  apiAddress: string;
};

export const useAqueductConsts = (): AqueductConsts => {
  return {
    apiAddress: process.env.SERVER_ADDRESS,
  };
};
