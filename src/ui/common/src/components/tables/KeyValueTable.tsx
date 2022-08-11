import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell, { TableCellProps } from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import * as React from 'react';

import { Data, DataSchema } from '../../utils/data';

// const dataSchema: DataSchema = {
//   fields: [
//     { name: 'Title', type: 'varchar' },
//     { name: 'Value', type: 'varchar' },
//   ],
//   pandas_version: '0.0.1', // TODO: Figure out what to set this value to.
// };

// function createData(): Data {
//   return {
//     schema: dataSchema,
//     data: [
//       ['avg_churn', '0.04'],
//       ['avg_workflows', '455'],
//       ['avg_users', '1.2'],
//       ['avg_users', '5'],
//     ],
//   };
// }

// const rows: Data = createData();

interface KeyValueTableProps {
  rows: Data;
  schema: DataSchema;
  height?: string;
  width?: string;
  maxHeight?: string;
  stickyHeader?: boolean;
  tableAlign?: string;
}

export const KeyValueTable: React.FC<KeyValueTableProps> = ({ rows, schema, height = "440px", width = "100%", maxHeight = "440px", stickyHeader = true, tableAlign = "left" }) => {
  return (
    <TableContainer sx={{ maxHeight, height, width }}>
      <Table stickyHeader={stickyHeader} aria-label={stickyHeader ? "sticky table" : "table"}>
        <TableHead>
          <TableRow>
            {schema.fields.map((column, idx) => (
              <TableCell
                key={`${column.name}-heading-${idx}`}
                align={tableAlign as TableCellProps["align"]}
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
                  const value = row[columnIndex];

                  return (
                    <TableCell
                      key={`cell-${rowIndex}-${columnIndex}`}
                      align={tableAlign as TableCellProps["align"]}
                    >
                      {value}
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

export default KeyValueTable;
