import {
  faCheck,
  faCircleCheck,
  faCode,
  faDatabase,
  faFileCode,
  faFileText,
  faHashtag,
  faImage,
  faList,
  faPencil,
  faSliders,
  faTableColumns,
  faTemperatureHalf,
} from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Tooltip } from '@mui/material';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React from 'react';
import { useSelector } from 'react-redux';
import { Handle, Position } from 'reactflow';

import { ReactFlowNodeData } from '../../../positioning/positioning';
import { RootState } from '../../../stores/store';
import { theme } from '../../../styles/theme/theme';
import { ArtifactType } from '../../../utils/artifacts';
import { OperatorType } from '../../../utils/operators';
import ExecutionStatus, { ExecState, FailureType } from '../../../utils/shared';
import ResourceItem from '../../pages/workflows/components/ResourceItem';
import { StatusIndicator } from '../workflowStatus';
import { BaseNode } from './BaseNode.styles';

const artifactNodeStatusLabels = {
  [ExecutionStatus.Succeeded]: 'Created',
  [ExecutionStatus.Failed]: 'Failed',
  [ExecutionStatus.Pending]: 'Pending',
  [ExecutionStatus.Canceled]: 'Canceled',
  [ExecutionStatus.Registered]: 'Registered',
  [ExecutionStatus.Running]: 'Running',
  [ExecutionStatus.Warning]: 'Warning',
  [ExecutionStatus.Unknown]: 'Unknown',
};

const operatorNodeStatusLabels = {
  [ExecutionStatus.Succeeded]: 'Succeeded',
  [ExecutionStatus.Failed]: 'Errored',
  [ExecutionStatus.Pending]: 'Pending',
  [ExecutionStatus.Canceled]: 'Canceled',
  [ExecutionStatus.Registered]: 'Registered',
  [ExecutionStatus.Running]: 'Running',
  [ExecutionStatus.Warning]: 'Warning',
  [ExecutionStatus.Unknown]: 'Unknown',
};

const checkNodeStatusLabels = {
  [ExecutionStatus.Succeeded]: 'Passed',
  [ExecutionStatus.Failed]: 'Failed',
  [ExecutionStatus.Pending]: 'Pending',
  [ExecutionStatus.Canceled]: 'Canceled',
  [ExecutionStatus.Registered]: 'Registered',
  [ExecutionStatus.Running]: 'Running',
  [ExecutionStatus.Warning]: 'Warning',
  [ExecutionStatus.Unknown]: 'Unknown',
};

export const artifactTypeToIconMapping = {
  [ArtifactType.String]: faFileText,
  [ArtifactType.Bool]: faCircleCheck,
  [ArtifactType.Numeric]: faHashtag,
  [ArtifactType.Dict]: faFileCode,
  // TODO: figure out if we should use other icon for tuple
  [ArtifactType.Tuple]: faFileCode,
  [ArtifactType.List]: faList,
  [ArtifactType.Table]: faTableColumns,
  [ArtifactType.Json]: faPencil,
  // TODO: figure out what to show for bytes.
  [ArtifactType.Bytes]: faFileCode,
  [ArtifactType.Image]: faImage,
  // TODO: Figure out what to show for Picklable
  [ArtifactType.Picklable]: faFileCode,
  [ArtifactType.Untyped]: faPencil,
};

export const operatorTypeToIconMapping = {
  [OperatorType.Param]: faSliders,
  [OperatorType.Function]: faCode,
  [OperatorType.Extract]: faDatabase,
  [OperatorType.Load]: faDatabase,
  [OperatorType.Metric]: faTemperatureHalf,
  [OperatorType.Check]: faCheck,
  [OperatorType.SystemMetric]: faTemperatureHalf,
};

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
  isConnectable: boolean;
};

const iconFontSize = '32px';

export const Node: React.FC<Props> = ({ data, isConnectable }) => {
  const selectedNodeId = useSelector(
    (state: RootState) =>
      state.workflowPageReducer.perWorkflowPageStates[data.dag.workflow_id]
        ?.SelectedNode?.nodeId
  );
  const isSelected = selectedNodeId === data.nodeId;
  const operatorType = data.operator?.spec?.type;
  const label =
    data.nodeType == 'operators' ? data.operator?.name : data.artifact?.name;
  let statusLabels;
  if (data.nodeType === 'artifacts') {
    statusLabels = artifactNodeStatusLabels;
  } else if (
    data.nodeType === 'operators' &&
    operatorType === OperatorType.Check
  ) {
    statusLabels = checkNodeStatusLabels;
  } else {
    // All other operators.
    statusLabels = operatorNodeStatusLabels;
  }

  // This is loaded at the top level of the workflow details page.
  const resourcesState = useSelector(
    (state: RootState) => state.resourcesReducer
  );

  let execState: ExecState;
  if (data.nodeType === 'operators') {
    execState = data.operatorResult?.exec_state;
  } else {
    execState = data.artifactResult?.exec_state;
  }

  const textColor = isSelected
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
        {!!data.artifactResult?.content_serialized &&
          data.operator?.spec?.type === OperatorType.Metric &&
          status === ExecutionStatus.Succeeded
          ? parseMetricResult(data.artifactResult?.content_serialized, 3)
          : statusLabels[status]}
      </Typography>
    </Box>
  );

  // Based on whether this is an operator or an artifact, select what icon to show in the header.
  let headerIcon;
  if (data.nodeType === 'operators') {
    if (
      operatorType === OperatorType.Extract ||
      operatorType === OperatorType.Load
    ) {
      const spec = data.operator?.spec?.extract ?? data.operator?.spec?.load; // One of these two must be set.
      if (!!spec) {
        headerIcon = (
          <ResourceItem
            resource={spec.service}
            resourceCustomName={
              resourcesState.resources[spec.resource_id]?.name
            }
            size={iconFontSize}
            defaultBackgroundColor={theme.palette.gray[200]}
            collapseName
          />
        );
      }
    } else {
      const engineSpec =
        data.operator?.spec?.engine_config ?? data.dag.engine_config;
      const resourceConfig = engineSpec[`${engineSpec.type}_config`];

      headerIcon = (
        <ResourceItem
          resource={engineSpec.type}
          resourceCustomName={
            resourcesState.resources[resourceConfig?.['resource_id']]?.name
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
        icon={artifactTypeToIconMapping[data.artifact?.type]}
        fontSize={iconFontSize}
      />
    );
  }

  // This is only used for metrics and checks to signify extra detail.
  let headerEndIcon;
  if (
    data.nodeType === 'operators' &&
    (operatorType === OperatorType.Check ||
      operatorType === OperatorType.Metric)
  ) {
    headerEndIcon = (
      <FontAwesomeIcon
        icon={operatorTypeToIconMapping[operatorType]}
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

            <Typography
              sx={{
                maxWidth: '70%',
                flex: 1,
                whiteSpace: 'nowrap',
                overflow: 'hidden',
                textOverflow: 'ellipsis',
                fontSize: '32px',
              }}
            >
              {label}
            </Typography>

            {headerEndIcon && (
              <>
                <Box justifySelf="end" ml={1} mr={2}>
                  <Tooltip
                    title={
                      operatorType === OperatorType.Check ? 'Check' : 'Metric'
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
