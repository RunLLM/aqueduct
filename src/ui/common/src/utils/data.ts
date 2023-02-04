import { Service } from './integrations';
import ExecutionStatus from './shared';

export const DataColumnTypeNames = [
  'varchar',
  'integer',
  'float',
  'boolean',
  'datetime',
  'json',
  'object',
] as const;

export type DataColumnType = typeof DataColumnTypeNames[number];

export type DataColumn = {
  /**
   * Name of column (key of object)
   */
  name: string;
  /**
   * Used to show an alternate text in column header.
   * e.g. colum named created_at but we wish to render as Created At
   */
  displayName?: string;
  /**
   * Type of data to be rendered in column.
   */
  type: DataColumnType;
};

export type DataSchema = {
  fields: DataColumn[];
  pandas_version: string;
};

export type TableRow = { [key: string]: string | number | boolean };
export type Data = {
  schema?: DataSchema;
  // data is an array of objects whose keys correspond to the schema above.
  // each element of the array corresponds to a row.
  // each key of the row object corresponds to a column.
  // column names must be unique (obviously ;) )
  data: TableRow[];
  status?: string;
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
  artifact_id: string;
  load_specs: DataPreviewLoadSpec[];
  versions: Record<string, DataPreviewVersion>;
};

export type DataPreview = {
  historical_versions: Record<string, DataPreviewInfo>;
  latest_versions: Record<string, DataPreviewInfo>;
};

export function inferSchema(
  rows: TableRow[],
  defaultType = 'object'
): DataSchema {
  if (!rows) {
    return { fields: [], pandas_version: '' };
  }

  return {
    fields: Object.keys(rows[0]).map((col) => ({
      name: col,
      displayName: col,
      type: defaultType,
    })),
    pandas_version: '',
  };
}
