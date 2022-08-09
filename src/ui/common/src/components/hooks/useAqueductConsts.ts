export type AqueductConsts = {
  apiAddress: string;
};

const DEFAULT_SERVER_ADDRESS = 'http://localhost:8080';

export const useAqueductConsts = (): AqueductConsts => {
  return {
    // Use default value for server address if there is not one set in the .env or env.local file.
    apiAddress: process.env.SERVER_ADDRESS ? process.env.SERVER_ADDRESS : DEFAULT_SERVER_ADDRESS,
  };
};
