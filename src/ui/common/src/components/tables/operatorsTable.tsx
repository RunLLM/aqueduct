import { Link } from '@mui/material';
import Box from '@mui/material/Box';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Typography from '@mui/material/Typography';
import React, { useState } from 'react';

import { OperatorsForResourceItem } from '../../reducers/resource';
import { getPathPrefix } from '../../utils/getPathPrefix';
import { ListWorkflowSummary } from '../../utils/workflows';
import OperatorParametersOverview from '../operators/operatorParametersOverview';
import { Button } from '../primitives/Button.styles';

type Props = {
  workflow?: ListWorkflowSummary;
  operators: OperatorsForResourceItem[];
};

const OperatorsTable: React.FC<Props> = ({ workflow, operators }) => {
  const [showInactive, setShowInactive] = useState(false);
  const shownOperators = showInactive
    ? operators
    : operators.filter((x) => x.is_active);
  const hasInactive = operators.filter((x) => !x.is_active).length;

  return (
    <TableContainer>
      <Table sx={{ minWidth: 650 }} aria-label="simple table">
        <TableHead>
          <TableRow>
            <TableCell align="left">
              <Typography variant="body2" color="gray.900">
                Operator
              </Typography>
            </TableCell>
            <TableCell align="left">
              <Typography variant="body2" color="gray.900">
                Type
              </Typography>
            </TableCell>
            <TableCell align="left">
              <Typography variant="body2" color="gray.900">
                Details
              </Typography>
            </TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {shownOperators.map((opInfo) => (
            <TableRow
              key={opInfo.operator.id}
              sx={{ '&:last-child td, &:last-child th': { border: 0 } }}
            >
              <TableCell align="left" scope="row">
                <Typography
                  variant="body2"
                  color={opInfo.is_active ? 'gray.800' : 'gray.600'}
                >
                  {opInfo.operator.name}
                </Typography>
              </TableCell>
              <TableCell align="left">
                <Typography
                  variant="body2"
                  color={opInfo.is_active ? 'gray.800' : 'gray.600'}
                >
                  {opInfo.operator.spec.type}
                </Typography>
              </TableCell>
              <TableCell align="left">
                <OperatorParametersOverview
                  operator={opInfo.operator}
                  textColor={opInfo.is_active ? 'gray.800' : 'gray.600'}
                />
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
      <Box flexDirection="row">
        {workflow && (
          <Link
            underline="none"
            href={`${getPathPrefix()}/workflow/${workflow.id}`}
          >
            <Button
              color="primary"
              sx={{ marginTop: '6px', marginRight: '8px' }}
            >
              {'Go to workflow details'}
            </Button>
          </Link>
        )}
        {/* This !! is necessary. Otherwise it becomes bitwise & op for integer. */}
        {!!hasInactive && (
          <Button
            color="secondary"
            sx={{ marginTop: '6px' }}
            onClick={() => setShowInactive(!showInactive)}
          >
            {showInactive
              ? 'Hide operators from previous versions'
              : 'Show operators from previous versions'}
          </Button>
        )}
      </Box>
    </TableContainer>
  );
};

export default OperatorsTable;
