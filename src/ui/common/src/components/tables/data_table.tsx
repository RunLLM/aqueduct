import Box from '@mui/material/Box';
//import { styled } from '@mui/material/styles';
//import Table from '@mui/material/Table';
//import TableBody from '@mui/material/TableBody';
//import TableCell, { tableCellClasses } from '@mui/material/TableCell';
//import TableHead from '@mui/material/TableHead';
//import TableRow from '@mui/material/TableRow';
import React from 'react';
import { Profiler } from "react";
import {AutoSizer, Table, Column} from 'react-virtualized';

import { Data, DataColumn } from '../../utils/data';

const logTimes = (id, phase, actualTime, baseTime, startTime, commitTime) => {
  console.log(`${id}'s ${phase} phase:`);
  console.log(`Actual time: ${actualTime}`);
  console.log(`Base time: ${baseTime}`);
  console.log(`Start time: ${startTime}`);
  console.log(`Commit time: ${commitTime}`);
};

type Props = {
  data: Data;
  width?: string;
};

/*function renderCell(
  key: number,
  column: DataColumn,
  value: string | number | boolean
) {
  // For now we just do plain rendering for all data types.
  return (
    <TableCell key={'cell-' + key}>
      {typeof value === 'boolean' ? value.toString() : value}
    </TableCell>
  );
}*/

/*const TintedTableRow = styled(TableRow)({
  '&:nth-of-type(odd)': {
    backgroundColor: 'white',
  },
  '&:nth-of-type(even)': {
    backgroundColor: 'gray50',
  },
  color: 'darkGray',
});*/

const DataTable: React.FC<Props> = ({ data, width }) => {
  /*const tableHeaderClasses = {
    [`&.${tableCellClasses.head}`]: {
      fontFamily: 'monospace',
      backgroundColor: 'blue.900',
      color: 'white',
    },
  };*/

  const columnSchema = data.schema.fields;
  const headers = columnSchema.map((column, idx) => {
    return (
      <Column label={column.name} dataKey={'header-' + idx} width={100} />
    );
  });

  //console.log(data.data)
  const sliced = data.data.slice(0, 100);

  /*const body = sliced.map((row, rowIdx) => {
    return (
      <TintedTableRow key={'row-' + rowIdx}>
        {Object.keys(row).map((value, idx) => {
          return renderCell(idx, columnSchema[idx], row[value]);
        })}
      </TintedTableRow>
    );
  });*/

  return (
    <Profiler id="DataTable" onRender={logTimes}>
      <Box
        sx={{
          overflow: 'auto',
          maxHeight: '100%',
          width: { width: width ? width : 'fit-content' },
          maxWidth: '100%',
        }}
      >
        <AutoSizer>
          {({height, width}) => (
            <Table
            width={width}
            height={height}
            headerHeight={20}
            rowHeight={30}
            rowCount={sliced.length}
            rowGetter={({index}) => sliced[index]}>
            {headers}
          </Table>
          )}
        </AutoSizer>,
      </Box>
    </Profiler>
  );
};

export default React.memo(DataTable);
