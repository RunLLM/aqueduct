import { faEllipsis, faMicrochip } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Alert, Collapse, Tooltip, Typography } from '@mui/material';
import Box from '@mui/material/Box';
import React, { useLayoutEffect, useState } from 'react';
import { useSelector } from 'react-redux';

import {
  useDagGetQuery,
  useDagResultsGetQuery,
  useNodesGetQuery,
  useWorkflowGetQuery,
} from '../../handlers/AqueductApi';
import { RootState } from '../../stores/store';
import { theme } from '../../styles/theme/theme';
import { EngineType } from '../../utils/engine';
import ExecutionStatus from '../../utils/shared';
import { reduceEngineTypes } from '../../utils/workflows';
import ResourceItem from '../pages/workflows/components/ResourceItem';
import WorkflowDescription from './WorkflowDescription';
import WorkflowNextUpdateTime from './WorkflowNextUpdateTime';
import { StatusIndicator } from './workflowStatus';
import VersionSelector from './WorkflowVersionSelector';

export const WorkflowPageContentId = 'workflow-page-main';

type Props = {
  apiKey: string;
  workflowId: string;
  dagId: string;
  dagResultId?: string;
};

const ContainerWidthBreakpoint = 700;

const WorkflowHeader: React.FC<Props> = ({
  apiKey,
  workflowId,
  dagId,
  dagResultId,
}) => {
  // TODO: Refactor
  const currentNode = useSelector(
    (state: RootState) => state.nodeSelectionReducer.selected
  );

  // NOTE: The 1000 here is just a placeholder. By the time the page snaps into place,
  // it will be overridden.
  const [containerWidth, setContainerWidth] = useState(1000);
  const narrowView = containerWidth < ContainerWidthBreakpoint;

  const getContainerSize = () => {
    const container = document.getElementById(WorkflowPageContentId);

    if (!container) {
      // The page hasn't fully rendered yet.
      setContainerWidth(1000); // Just a default value.
    } else {
      setContainerWidth(container.clientWidth);
    }
  };

  // TODO (ENG-2302): useLayoutEffect here. May want to figure out some way to debounce as it gets called quite quickly when resizing.
  window.addEventListener('resize', getContainerSize);
  useLayoutEffect(getContainerSize, [currentNode]);

  const [showDescription, setShowDescription] = useState(false);
  const { data: workflow } = useWorkflowGetQuery({ apiKey, workflowId });
  const { data: dag } = useDagGetQuery({ apiKey, workflowId, dagId });
  const { data: dagResults } = useDagResultsGetQuery({ apiKey, workflowId });
  const { data: nodes } = useNodesGetQuery({ apiKey, workflowId, dagId });
  const dagResult = dagResults.filter((x) => x.dag_id === dagResultId)[0];

  if (!dag || !workflow || !nodes) {
    // We simply do not render if main data is not available.
    // We expect caller to handle loading and error status.
    return null;
  }

  const status = dagResult?.exec_state?.status ?? ExecutionStatus.Unknown;
  const name = workflow.name;

  const showAirflowUpdateWarning =
    dag.engine_config.type === EngineType.Airflow &&
    !dag.engine_config.airflow_config?.matches_airflow;
  const airflowUpdateWarning = (
    <Box maxWidth="800px">
      <Alert severity="warning">
        Please copy the latest Airflow DAG file to your Airflow server if you
        have not done so already. New Airflow DAG runs will not be synced
        properly with Aqueduct until you have copied the file.
      </Alert>
    </Box>
  );

  const engines = reduceEngineTypes(
    dag.engine_config.type,
    nodes.operators.map((op) => op.spec?.engine_config?.type).filter((t) => !!t)
  );

  return (
    <Box>
      <Box
        sx={{
          display: 'flex',
          alignItems: narrowView ? 'start' : 'center',
          flexDirection: narrowView ? 'column' : 'row',
          transition: 'height 100ms',
        }}
      >
        <Box sx={{ display: 'flex', alignItems: 'center' }}>
          {!!dagResult && <StatusIndicator status={status} />}

          <Typography
            variant="h5"
            sx={{ ml: 1, lineHeight: 1 }}
            fontWeight="normal"
          >
            {name}
          </Typography>

          <Box ml={2}>
            <VersionSelector apiKey={apiKey} workflowId={workflowId} />
          </Box>
        </Box>

        <Box
          sx={{
            fontSize: '20px',
            p: 1,
            ml: narrowView ? 0 : 1,
            mt: narrowView ? 1 : 0,
            borderRadius: '8px',
            ':hover': {
              backgroundColor: theme.palette.gray[50],
            },
            cursor: 'pointer',
          }}
          onClick={() => setShowDescription(!showDescription)}
        >
          <Tooltip title="See more" arrow>
            <FontAwesomeIcon
              icon={faEllipsis}
              style={{
                transform: showDescription ? 'rotateX(180deg)' : '',
                transition: 'transform 200ms',
              }}
            />
          </Tooltip>
        </Box>
      </Box>

      <Collapse in={showDescription}>
        <Box display="flex" alignItems="center" my={1}>
          {/* Display the Workflow Engine. */}
          <Tooltip title={'Compute Engine(s)'} arrow>
            <Box display="flex" alignItems="center">
              <FontAwesomeIcon
                icon={faMicrochip}
                color={theme.palette.gray[800]}
              />
              <Box display="flex" flexDirection="row">
                {engines.map((engine) => (
                  <Box ml={1} key={engine}>
                    <ResourceItem resource={engine} />
                  </Box>
                ))}
              </Box>
            </Box>
          </Tooltip>
          <WorkflowNextUpdateTime workflow={workflow} />
        </Box>

        <Box
          sx={{
            backgroundColor: theme.palette.gray[25],
            px: 2,
            py: 1,
            my: 1,
            maxWidth: '800px',
            borderRadius: '4px',
          }}
        >
          <WorkflowDescription workflow={workflow} />
        </Box>
      </Collapse>

      {showAirflowUpdateWarning && airflowUpdateWarning}
    </Box>
  );
};

export default WorkflowHeader;
