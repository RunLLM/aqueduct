import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React from 'react';

import {
  NodeResultsMap,
  OperatorResponse,
} from '../../../handlers/responses/node';
import { DataSchema } from '../../../utils/data';
import { CheckLevel } from '../../../utils/operators';
import ExecutionStatus from '../../../utils/shared';
import OperatorExecStateTable, {
  OperatorExecStateTableType,
} from '../../tables/OperatorExecStateTable';

// This file contains two major components to show metrics and checks
// associated with a given artifact.

const schema: DataSchema = {
  fields: [
    { name: 'Title', type: 'varchar' },
    { name: 'Value', type: 'varchar' },
  ],
  pandas_version: '',
};

type MetricsOverviewProps = {
  nodeResults?: NodeResultsMap;
  metrics: OperatorResponse[];
};

export const MetricsOverview: React.FC<MetricsOverviewProps> = ({
  nodeResults,
  metrics,
}) => {
  const metricTableEntries = {
    schema: schema,
    data: metrics.map((metricOp) => {
      const title = metricOp.name;
      const opResult = (nodeResults?.operators ?? {})[metricOp.id];
      const artfResult = (nodeResults?.artifacts ?? {})[metricOp.outputs[0]];
      const value = artfResult?.content_serialized;
      const status = opResult?.exec_state?.status;

      return {
        title,
        value,
        status,
      };
    }),
  };

  return (
    <Box width="100%">
      <Typography
        variant="h6"
        component="div"
        marginBottom="8px"
        fontWeight="normal"
      >
        Metrics
      </Typography>
      {metricTableEntries.data.length > 0 ? (
        <OperatorExecStateTable
          schema={metricTableEntries.schema}
          rows={metricTableEntries}
          tableType={OperatorExecStateTableType.Metric}
        />
      ) : (
        <Typography variant="body2" color="gray.700">
          This artifact has no Metrics.
        </Typography>
      )}
    </Box>
  );
};

export type ChecksOverviewProps = {
  nodeResults?: NodeResultsMap;
  checks: OperatorResponse[];
};

export const ChecksOverview: React.FC<ChecksOverviewProps> = ({
  nodeResults,
  checks,
}) => {
  const checkTableEntries = {
    schema: schema,
    data: checks.map((checkOp) => {
      const name = checkOp.name;
      const opResult = (nodeResults?.operators ?? {})[checkOp.id];
      let status = opResult?.exec_state?.status;
      if (
        status === ExecutionStatus.Failed &&
        checkOp.spec?.check?.level === CheckLevel.Warning
      ) {
        status = ExecutionStatus.Warning;
      }

      return {
        title: name,
        value: status,
        status: status,
      };
    }),
  };

  return (
    <Box width="100%">
      <Typography
        variant="h6"
        component="div"
        marginBottom="8px"
        fontWeight="normal"
      >
        Checks
      </Typography>
      {checkTableEntries.data.length > 0 ? (
        <OperatorExecStateTable
          schema={checkTableEntries.schema}
          rows={checkTableEntries}
          tableType={OperatorExecStateTableType.Check}
        />
      ) : (
        <Typography variant="body2" color="gray.700">
          This artifact has no Checks.
        </Typography>
      )}
    </Box>
  );
};
