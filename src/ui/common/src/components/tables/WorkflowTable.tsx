import Box from '@mui/material/Box';
import Paper from '@mui/material/Paper';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TablePagination from '@mui/material/TablePagination';
import TableRow from '@mui/material/TableRow';
import Typography from '@mui/material/Typography';
import * as React from 'react';

import { DataSchema } from '../../utils/data';

export type WorkflowTableRow = { [key: string]: string | number | boolean | JSX.Element };

export interface WorkflowTableData {
    schema?: DataSchema,
    data: WorkflowTableRow[]
}

export interface WorkflowsTableProps {
    data: WorkflowTableData;
}

export const WorkflowTable: React.FC<WorkflowsTableProps> = ({ data }) => {
    const [page, setPage] = React.useState(0);
    const [rowsPerPage, setRowsPerPage] = React.useState(5);

    const rows = data.data;
    const columns = data.schema.fields;

    const handleChangePage = (event: unknown, newPage: number) => {
        setPage(newPage);
    };

    const handleChangeRowsPerPage = (
        event: React.ChangeEvent<HTMLInputElement>
    ) => {
        setRowsPerPage(+event.target.value);
        setPage(0);
    };

    return (
        <Paper sx={{ overflow: 'hidden' }}>
            <TableContainer sx={{ maxHeight: '400px' }}>
                <Table stickyHeader aria-label="sticky table">
                    <TableHead>
                        <TableRow>
                            {columns.map((column, columnIndex) => {
                                return (
                                    <TableCell
                                        key={`table-header-col-${columnIndex}`}
                                        align={'left'}
                                        sx={{
                                            backgroundColor: 'blue.900',
                                            color: 'white',
                                            minWidth: '80px',
                                        }}
                                    >
                                        <Box flexDirection="column">
                                            <Typography
                                                variant="body1"
                                                sx={{
                                                    textTransform: 'none',
                                                    fontFamily: 'monospace',
                                                    fontSize: '16px',
                                                }}
                                            >
                                                {column.name}
                                            </Typography>
                                        </Box>
                                    </TableCell>
                                );
                            })}
                        </TableRow>
                    </TableHead>
                    <TableBody>
                        {rows
                            .slice(page * rowsPerPage, page * rowsPerPage + rowsPerPage)
                            .map((row, rowIndex) => {
                                return (
                                    <TableRow
                                        hover
                                        role="checkbox"
                                        tabIndex={-1}
                                        key={`table-row-${rowIndex}`}
                                    >
                                        {columns.map((column, columnIndex) => {
                                            const value = row[column.name];
                                            return (
                                                <TableCell
                                                    key={`table-col-${columnIndex}`}
                                                    align={'left'}
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
            <TablePagination
                rowsPerPageOptions={[5, 10, 25, 50, 100]}
                component="div"
                count={rows.length}
                rowsPerPage={rowsPerPage}
                page={page}
                onPageChange={handleChangePage}
                onRowsPerPageChange={handleChangeRowsPerPage}
            />
        </Paper>
    );
};

export default WorkflowTable;
