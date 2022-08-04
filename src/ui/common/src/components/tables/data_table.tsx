import Box from '@mui/material/Box';
//import { styled } from '@mui/material/styles';
//import Table from '@mui/material/Table';
//import TableBody from '@mui/material/TableBody';
import TableCell, { tableCellClasses } from '@mui/material/TableCell';
//import TableHead from '@mui/material/TableHead';
//import TableRow from '@mui/material/TableRow';
import React from 'react';
import { Profiler } from "react";
import {AutoSizer, Table, Column} from 'react-virtualized';
import 'react-virtualized/styles.css'; // only needs to be imported once

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
  const tableHeaderClasses = {
    [`&.${tableCellClasses.head}`]: {
      fontFamily: 'monospace',
      backgroundColor: 'blue.900',
      color: 'white',
    },
  };

  function customHeaderRenderer({
    columnData,
    dataKey,
    disableSort,
    label,
    sortBy,
    sortDirection,
  }) {
    console.log(label.name);
    console.log(label.type);
    return (
      <TableCell sx={tableHeaderClasses} >
        <span style={{ fontSize: '16px' }}>{label.name}</span> <br />{' '}
        <span style={{ fontSize: '12px' }}> {label.type} </span>
      </TableCell>
    )
  }

  const columnSchema = data.schema.fields;
  const headers = columnSchema.map((column, idx) => {
    return (
      <Column label={column} dataKey={column.name} width={50} flexGrow={1} headerRenderer={customHeaderRenderer} />
    );
  });

  //console.log(data.data)
  //const sliced = data.data.slice(0, 100);

  /*const body = sliced.map((row, rowIdx) => {
    return (
      <TintedTableRow key={'row-' + rowIdx}>
        {Object.keys(row).map((value, idx) => {
          return renderCell(idx, columnSchema[idx], row[value]);
        })}
      </TintedTableRow>
    );
  });*/

  const MIN_TABLE_WIDTH = 100 * columnSchema.length;

  return (
    <Profiler id="DataTable" onRender={logTimes}>
      <AutoSizer>
        {({height, width}) => (
          <Table
          headerStyle={tableHeaderClasses}
          width={width < MIN_TABLE_WIDTH ? MIN_TABLE_WIDTH : width}
          height={height}
          headerHeight={100}
          rowHeight={30}
          rowCount={data.data.length}
          rowGetter={({index}) => {
            return data.data[index]
          }}>
          {headers}
        </Table>
        )}
      </AutoSizer>
    </Profiler>
  );
};

export default React.memo(DataTable);
