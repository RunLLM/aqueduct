import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableFooter from '@mui/material/TableFooter';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Typography from '@mui/material/Typography';
import React, { useState } from 'react';

import { OperatorsForIntegrationItem } from '../../reducers/integrationOperators';

type Props = {
  operators: OperatorsForIntegrationItem[];
};

const OperatorsTable: React.FC<Props> = ({ operators }) => {
  const [showInactive, setShowInactive] = useState(false);
  const shownOperators = showInactive
    ? operators
    : operators.filter((x) => x.is_active);

  return (
    <TableContainer>
      <Table sx={{ minWidth: 650 }} aria-label="simple table">
        <TableHead>
          <TableRow>
            <TableCell>
              {' '}
              <Typography variant="body2" color="gray.900">
                Operator{' '}
              </Typography>
            </TableCell>
            <TableCell align="right">
              <Typography
                variant="body2"
                color="gray.900"
                onClick={() => setShowInactive(!showInactive)}
              >
                Type
              </Typography>
            </TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {shownOperators.map((opInfo) => (
            <TableRow key={opInfo.operator.id}>
              <TableCell align="left" scope="row">
                <Typography variant="body2" color="gray.800">
                  {opInfo.operator.name}
                </Typography>
              </TableCell>
              <TableCell align="right">
                <Typography variant="body2" color="gray.800">
                  {opInfo.operator.spec.type}
                </Typography>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
        <TableFooter sx={{ marginTop: '2px' }}>
          <TableRow>
            <Typography
              variant="body2"
              color="gray.800"
              onClick={() => setShowInactive(!showInactive)}
            >
              {showInactive
                ? 'only show operators from current version'
                : 'show operators from older versions'}
            </Typography>
          </TableRow>
        </TableFooter>
      </Table>
    </TableContainer>
  );
};

export default OperatorsTable;
