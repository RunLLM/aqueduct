import {
  faCircleDown,
  faUpRightAndDownLeftFromCenter,
} from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Box, Tooltip } from '@mui/material';
import React from 'react';
import { useNavigate } from 'react-router-dom';

import {
  ArtifactResponse,
  OperatorResponse,
} from '../../handlers/responses/node';
import { NodeSelection } from '../../reducers/pages/Workflow';
import UserProfile from '../../utils/auth';
import { handleExportFunction, OperatorType } from '../../utils/operators';
import { Button } from '../primitives/Button.styles';

type Props = {
  user: UserProfile;
  workflowId: string;
  dagResultId: string;
  selectedNodeState: NodeSelection;
  selectedNode: OperatorResponse | ArtifactResponse;
};

const WorkflowNodeSidesheetActions: React.FC<Props> = ({
  user,
  workflowId,
  dagResultId,
  selectedNodeState,
  selectedNode,
}) => {
  const navigate = useNavigate();
  const buttonStyle = {
    fontSize: '20px',
    mr: 1,
  };

  let navigateButton;
  let includeExportOpButton = true;

  if (!dagResultId) {
    return null;
  } else {
    let navigationUrl;
    if (selectedNodeState.nodeType === 'artifacts') {
      navigationUrl = `/workflow/${workflowId}/result/${dagResultId}/artifact/${selectedNodeState.nodeId}`;
      includeExportOpButton = false;
    } else {
      const opNode = selectedNode as OperatorResponse;
      if (opNode.spec?.type === OperatorType.Metric) {
        navigationUrl = `/workflow/${workflowId}/result/${dagResultId}/metric/${opNode.id}`;
      } else if (opNode.spec?.type === OperatorType.Check) {
        navigationUrl = `/workflow/${workflowId}/result/${dagResultId}/check/${opNode.id}`;
      } else {
        navigationUrl = `/workflow/${workflowId}/result/${dagResultId}/operator/${opNode.id}`;
        if (opNode.spec?.type !== OperatorType.Function) {
          // This is a load or save operator.
          includeExportOpButton = false;
        }
      }
    }

    navigateButton = (
      <Button
        variant="text"
        sx={buttonStyle}
        onClick={() => {
          navigate(navigationUrl);
        }}
      >
        <Tooltip title="Expand Details" arrow>
          <FontAwesomeIcon icon={faUpRightAndDownLeftFromCenter} />
        </Tooltip>
      </Button>
    );
  }

  const exportOpButton = (
    <Button
      onClick={async () => {
        await handleExportFunction(
          user,
          selectedNodeState.nodeId,
          `${selectedNode.name ?? 'function'}.zip`
        );
      }}
      variant="text"
      sx={buttonStyle}
    >
      <Tooltip title="Download Code" arrow>
        <FontAwesomeIcon icon={faCircleDown} />
      </Tooltip>
    </Button>
  );

  return (
    <Box display="flex" alignItems="center">
      {includeExportOpButton && exportOpButton}
      {navigateButton}
    </Box>
  );
};

export default WorkflowNodeSidesheetActions;
