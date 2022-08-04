import * as React from 'react';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import { Data, DataSchema } from '../../utils/data';

const dataSchema: DataSchema = {
    fields: [{ name: 'Title', type: 'varchar' }, { name: 'Value', type: 'varchar' }],
    pandas_version: '0.0.1' // TODO: Figure out what to set this value to. 
}

function createData(): Data {
    return {
        schema: dataSchema,
        data: [
            ['avg_churn', '0.04'],
            ['avg_workflows', '455'],
            ['avg_users', '1.2'],
            ['avg_users', '5']
        ]
    }
}

const rows: Data = createData();

interface KeyValueTableProps {
    data: Data
}

const tableAlign = "left";

export const KeyValueTable: React.FC = () => {
    return (
        <TableContainer sx={{ maxHeight: 440, height: 440, width: '100%' }}>
            <Table stickyHeader aria-label="sticky table">
                <TableHead>
                    <TableRow>
                        {dataSchema.fields.map((column, idx) => (
                            <TableCell
                                key={`${column.name}-heading-${idx}`}
                                align={tableAlign}
                                style={{}}
                            >
                                {column.name}
                            </TableCell>
                        ))}
                    </TableRow>
                </TableHead>
                <TableBody>
                    {rows.data.map((row, rowIndex) => {
                        return (
                            <TableRow hover role="checkbox" tabIndex={-1} key={`tableBody-${rowIndex}`}>
                                {
                                    dataSchema.fields.map((column, columnIndex) => {
                                        const value = row[columnIndex];

                                        // TODO: Consider making cell alignment a property to pass in.
                                        return (
                                            <TableCell key={`cell-${rowIndex}-${columnIndex}`} align={tableAlign}>
                                                {value}
                                            </TableCell>
                                        )
                                    })
                                }
                            </TableRow>
                        )
                    })
                    }
                </TableBody>
            </Table>
        </TableContainer>
    );
}

export default KeyValueTable;