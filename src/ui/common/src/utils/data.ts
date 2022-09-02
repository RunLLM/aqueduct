import { Service } from './integrations';
import ExecutionStatus from './shared';

export const DataColumnTypeNames = [
  'varchar',
  'integer',
  'float',
  'boolean',
  'datetime',
  'json',
] as const;

export type DataColumnType = typeof DataColumnTypeNames[number];

export type DataColumn = { name: string; type: DataColumnType };

export type DataSchema = {
  fields: DataColumn[];
  pandas_version: string;
};

export type Data = {
  schema?: DataSchema;
  //data: (string | number | boolean)[][];
  // data is an array of objects whose keys correspond to the schema above.
  // each record of data is a row on the table.
  data: any[]
};

export type DataPreviewLoadSpec = {
  service: Service;
  integration_id: string;
  parameters: Record<string, string>;
};

export type DataPreviewVersion = {
  error: string;
  status: ExecutionStatus;
  timestamp: number;
};

export type DataPreviewInfo = {
  workflow_name: string;
  workflow_id: string;
  artifact_name: string;
  load_specs: DataPreviewLoadSpec[];
  versions: Record<string, DataPreviewVersion>;
};

export type DataPreview = {
  historical_versions: Record<string, DataPreviewInfo>;
  latest_versions: Record<string, DataPreviewInfo>;
};
