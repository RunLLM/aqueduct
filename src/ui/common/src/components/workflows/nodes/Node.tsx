import { IconDefinition } from '@fortawesome/fontawesome-svg-core';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Tooltip } from '@mui/material';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React from 'react';
import { useSelector } from 'react-redux';
import { Handle, Position } from 'reactflow';

import { RootState } from '../../../stores/store';
import { theme } from '../../../styles/theme/theme';
import { OperatorType } from '../../../utils/operators';
import { ReactFlowNodeData, ReactflowNodeType } from '../../../utils/reactflow';
import ExecutionStatus, { ExecState, FailureType } from '../../../utils/shared';
import ResourceItem from '../../pages/workflows/components/ResourceItem';
import { StatusIndicator } from '../workflowStatus';
import { BaseNode } from './BaseNode.styles';
import {
  artifactNodeStatusLabels,
  artifactTypeToIconMapping,
  checkNodeStatusLabels,
  operatorNodeStatusLabels,
  operatorTypeToIconMapping,
} from './nodeTypes';

export const parseMetricResult = (
  metricValue: string,
  sigfigs: number
): string => {
  // Check if the number passed in is a whole number, return that if so.
  const parsedFloat = parseFloat(metricValue);
  if (parsedFloat % 1 === 0) {
    return metricValue;
  }
  // Only show three decimal points.
  return parsedFloat.toFixed(sigfigs);
};

type Props = {
  data: ReactFlowNodeData;
  defaultLabel: string;
  isConnectable: boolean;
  icon?: IconDefinition;
  statusLabels: { [key: string]: string };
  // The preview is only shown if the status of this node is succeeded.
  // If it is, then we replace the label with the preview. If the preview
  // is null or the status is not succeeded, then we show the regular label.
  preview?: string;
};

const iconFontSize = '32px';

export const Node: React.FC<Props> = ({ data, isConnectable }) => {
  const currentNode = useSelector(
    (state: RootState) => state.nodeSelectionReducer.selected
  );
  const workflowState = useSelector(
    (state: RootState) => state.workflowReducer
  );

  let statusLabels;
  if (data.nodeType === ReactflowNodeType.Artifact) {
    statusLabels = artifactNodeStatusLabels;
  } else if (
    data.nodeType === ReactflowNodeType.Operator &&
    data.spec.type === OperatorType.Check
  ) {
    statusLabels = checkNodeStatusLabels;
  } else {
    // All other operators.
    statusLabels = operatorNodeStatusLabels;
  }

  // This is loaded at the top level of the workflow details page.
  const integrationsState = useSelector(
    (state: RootState) => state.integrationsReducer
  );

  const selected = currentNode.id === data.nodeId;

  let execState: ExecState;
  if (data.nodeType === ReactflowNodeType.Operator) {
    execState = workflowState.operatorResults[data.nodeId]?.result?.exec_state;
  } else {
    execState = workflowState.artifactResults[data.nodeId]?.result?.exec_state;
  }

  const textColor = selected
    ? theme.palette.DarkContrast50
    : theme.palette.DarkContrast;
  const borderColor = textColor;

  let status = execState?.status;
  if (
    execState?.status === ExecutionStatus.Failed &&
    execState.failure_type == FailureType.UserNonFatal
  ) {
    status = ExecutionStatus.Warning;
  }

  let backgroundColor;
  switch (status) {
    case ExecutionStatus.Succeeded:
      backgroundColor = theme.palette.green[100];
      break;
    case ExecutionStatus.Warning:
      backgroundColor = theme.palette.yellow[100];
      break;
    case ExecutionStatus.Failed:
      backgroundColor = theme.palette.red[100];
      break;
    case ExecutionStatus.Canceled:
    case ExecutionStatus.Pending:
    default:
      backgroundColor = theme.palette.gray[400];
  }

  const statusIndicatorComponent = !!execState?.status && (
    <Box
      sx={{
        display: 'flex',
        alignItems: 'center',
        backgroundColor: backgroundColor,
        // Even though the BaseNode's border radius is 8px, the differing dimensions
        // make the same node width look funny. We set it at 5px to get rid of any
        // whitespace.
        borderBottomRightRadius: '5px',
        borderBottomLeftRadius: '5px',
      }}
      flex={1}
      height="50%"
      width="100%"
    >
      <Box ml={1}>
        <StatusIndicator
          status={status}
          size={iconFontSize}
          includeTooltip={false}
        />
      </Box>

      <Typography
        ml={1}
        textTransform="capitalize"
        fontSize="28px"
        fontWeight="light"
      >
        {/* Only show the preview if the status is succeeded and it exists. Otherwise,
         * show the label that we're given. The reason for this is (eg) for a metric,
         * if the status is either pending or failed/canceled/etc., the preview will be
         * NaN. This only applies to metric operators. */}
        {!!data.result &&
        data.spec?.type === OperatorType.Metric &&
        status === ExecutionStatus.Succeeded
          ? parseMetricResult(data.result, 3)
          : statusLabels[status]}
      </Typography>
    </Box>
  );

  // Based on whether this is an operator or an artifact, select what icon to show in the header.
  let headerIcon;
  if (data.nodeType === ReactflowNodeType.Operator) {
    if (
      data.spec.type === OperatorType.Extract ||
      data.spec.type === OperatorType.Load
    ) {
      const spec = data.spec.extract ?? data.spec.load; // One of these two must be set.
      headerIcon = (
        <ResourceItem
          resource={spec.service}
          resourceCustomName={
            integrationsState.integrations[spec.integration_id]?.name
          }
          size={iconFontSize}
          defaultBackgroundColor={theme.palette.gray[200]}
          collapseName
        />
      );
    } else {
      const engineSpec = data.spec?.engine_config ?? data.dagEngineConfig;
      const integrationConfig = engineSpec[`${engineSpec.type}_config`];

      headerIcon = (
        <ResourceItem
          resource={engineSpec.type}
          resourceCustomName={
            integrationsState.integrations[
              integrationConfig?.['integration_id']
            ]?.name
          }
          size={iconFontSize}
          defaultBackgroundColor={theme.palette.gray[200]}
          collapseName
        />
      );
    }
  } else {
    // This is an artifact.
    headerIcon = (
      <FontAwesomeIcon
        icon={artifactTypeToIconMapping[data.artifactType]}
        fontSize={iconFontSize}
      />
    );
  }

  // This is only used for metrics and checks to signify extra detail.
  let headerEndIcon;
  if (
    data.nodeType === ReactflowNodeType.Operator &&
    (data.spec?.type === OperatorType.Check ||
      data.spec?.type === OperatorType.Metric)
  ) {
    headerEndIcon = (
      <FontAwesomeIcon
        icon={operatorTypeToIconMapping[data.spec.type]}
        fontSize={iconFontSize}
      />
    );
  }

  return (
    <Box>
      <BaseNode
        sx={{
          color: textColor,
          borderColor: borderColor,
        }}
      >
        <Box
          display="flex"
          flexDirection="column"
          alignItems="start"
          width="100%"
          height="100%"
        >
          <Box
            display="flex"
            alignItems="center"
            width="100%"
            height="50%"
            flex={1}
            sx={{
              backgroundColor: theme.palette.gray[200],
              // Even though the BaseNode's border radius is 8px, the differing dimensions
              // make the same node width look funny. We set it at 5px to get rid of any
              // whitespace.
              borderTopLeftRadius: '5px',
              borderTopRightRadius: '5px',
            }}
          >
            <Box sx={{ ml: 1, mr: 2, fontSize: iconFontSize }}>
              {headerIcon}
            </Box>

            <Box flex={1}>
              <Typography
                sx={{
                  maxWidth: '80%',
                  flex: 1,
                  whiteSpace: 'nowrap',
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                  fontSize: '32px',
                }}
              >
                {data.label}
              </Typography>
            </Box>

            {headerEndIcon && (
              <>
                <Box justifySelf="end" mr={2}>
                  <Tooltip
                    title={
                      data.spec?.type === OperatorType.Check
                        ? 'Check'
                        : 'Metric'
                    }
                    arrow
                  >
                    {headerEndIcon}
                  </Tooltip>
                </Box>
              </>
            )}
          </Box>

          {statusIndicatorComponent}
        </Box>

        <Handle
          type="source"
          id="db-source-id"
          style={{
            background: theme.palette.DarkContrast,
            border: theme.palette.DarkContrast,
          }}
          isConnectable={isConnectable}
          position={Position.Right}
        />

        <Handle
          type="target"
          id="db-target-id"
          style={{
            background: theme.palette.DarkContrast,
            border: theme.palette.DarkContrast,
          }}
          isConnectable={isConnectable}
          position={Position.Left}
        />
      </BaseNode>
    </Box>
  );
};

export default Node;
