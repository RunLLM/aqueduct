import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell, { TableCellProps } from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Typography from '@mui/material/Typography';
import * as React from 'react';

import { Data, DataSchema } from '../../utils/data';
import { CheckTableItem } from './CheckTableItem';

export enum OperatorExecStateTableType {
  Metric = 'metric',
  Check = 'check',
}

interface OperatorExecStateTableProps {
  rows: Data;
  schema?: DataSchema;
  height?: string;
  width?: string;
  maxHeight?: string;
  stickyHeader?: boolean;
  tableAlign?: string;
  tableType: OperatorExecStateTableType;
}

const kvSchema: DataSchema = {
  fields: [
    { name: 'Title', type: 'varchar' },
    { name: 'Value', type: 'varchar' },
  ],
  pandas_version: '0.0.1', // TODO: Figure out what to set this value to.
};

export const OperatorExecStateTable: React.FC<OperatorExecStateTableProps> = ({
  rows,
  schema = kvSchema,
  height = '440px',
  width = '100%',
  maxHeight = '440px',
  stickyHeader = true,
  tableAlign = 'left',
  tableType,
}) => {
  return (
    <TableContainer sx={{ maxHeight, height, width }}>
      <Table
        stickyHeader={stickyHeader}
        aria-label={stickyHeader ? 'sticky table' : 'table'}
      >
        <TableHead>
          <TableRow>
            {schema.fields.map((column, idx) => (
              <TableCell
                key={`${column.name}-heading-${idx}`}
                align={tableAlign as TableCellProps['align']}
              >
                {column.name}
              </TableCell>
            ))}
          </TableRow>
        </TableHead>
        <TableBody>
          {rows.data.map((row, rowIndex) => {
            return (
              <TableRow
                hover
                role="checkbox"
                tabIndex={-1}
                key={`tableBody-${rowIndex}`}
              >
                {schema.fields.map((column, columnIndex) => {
                  const value = row[column.name.toLowerCase()];

                  // For title columns we should just render the text.
                  // For a check's value column, we should render the appropriate icon.
                  return (
                    <TableCell
                      key={`cell-${rowIndex}-${columnIndex}`}
                      align={tableAlign as TableCellProps['align']}
                    >
                      {tableType === OperatorExecStateTableType.Metric ||
                      column.name === 'title' ? (
                        <Typography
                          sx={{
                            fontWeight: column.name !== 'title' ? 300 : null,
                          }}
                        >
                          {value.toString()}
                        </Typography>
                      ) : (
                        <CheckTableItem checkValue={value as string} />
                      )}
                    </TableCell>
                  );
                })}
              </TableRow>
            );
          })}
        </TableBody>
      </Table>
    </TableContainer>
  );
};

export default OperatorExecStateTable;
