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
            <TableCell> Operator </TableCell>
            <TableCell align="right"> Type </TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {shownOperators.map((opInfo) => (
            <TableRow key={opInfo.operator.id}>
              <TableCell align="left" scope="row">
                {opInfo.operator.name}
              </TableCell>
              <TableCell align="right">{opInfo.operator.spec.type}</TableCell>
            </TableRow>
          ))}
        </TableBody>
        <TableFooter>
          <TableRow>
            <Typography
              variant="body2"
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
