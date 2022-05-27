import Box from '@mui/material/Box';
import { styled } from '@mui/material/styles';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell, { tableCellClasses } from '@mui/material/TableCell';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import React from 'react';
import {Data, DataColumn} from "../../utils/data";

type Props = {
    data: Data;
    width?: string;
};

function renderCell(key: number, column: DataColumn, value: string | number | boolean) {
    // For now we just do plain rendering for all data types.
    return <TableCell key={'cell-' + key}>{typeof value === 'boolean' ? value.toString() : value}</TableCell>;
}

const TintedTableRow = styled(TableRow)({
    '&:nth-of-type(odd)': {
        backgroundColor: 'white',
    },
    '&:nth-of-type(even)': {
        backgroundColor: 'gray50',
    },
    color: 'darkGray',
});

const DataTable: React.FC<Props> = ({ data, width }) => {
    const tableHeaderClasses = {
        [`&.${tableCellClasses.head}`]: {
            fontFamily: 'monospace',
            backgroundColor: 'blue.900',
            color: 'white',
        },
    };

    const columnSchema = data.schema.fields;
    const headers = columnSchema.map((column, idx) => {
        return (
            <TableCell sx={tableHeaderClasses} key={'header-' + idx}>
                <span style={{ fontSize: '16px' }}>{column.name}</span> <br />{' '}
                <span style={{ fontSize: '12px' }}> {column.type} </span>
            </TableCell>
        );
    });

    const body = data.data.map((row, rowIdx) => {
        return (
            <TintedTableRow key={'row-' + rowIdx}>
                {Object.keys(row).map((value, idx) => {
                    return renderCell(idx, columnSchema[idx], row[value]);
                })}
            </TintedTableRow>
        );
    });

    return (
        <Box
            sx={{
                overflow: 'auto',
                maxHeight: '100%',
                width: { width: width ? width : 'fit-content' },
                maxWidth: '100%',
            }}
        >
            <Table>
                <TableHead>
                    <TableRow>{headers}</TableRow>
                </TableHead>
                <TableBody>{body}</TableBody>
            </Table>
        </Box>
    );
};

export default React.memo(DataTable);
