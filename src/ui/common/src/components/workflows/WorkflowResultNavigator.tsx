import {
  faChevronLeft,
  faChevronRight,
} from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Box, Tooltip } from '@mui/material';
import React from 'react';
import { useNavigate } from 'react-router-dom';
import { useDagResultsGetQuery } from 'src/handlers/AqueductApi';

import { useWorkflowIds } from '../pages/workflow/id/hook';
import { Button } from '../primitives/Button.styles';

type Props = {
  apiKey: string;
};

const WorkflowResultNavigator: React.FC<Props> = ({ apiKey }) => {
  const navigate = useNavigate();
  const { workflowId, dagResultId } = useWorkflowIds(apiKey);
  const { data: dagResults } = useDagResultsGetQuery(
    { apiKey, workflowId },
    { skip: !workflowId || !dagResultId }
  );

  if (!dagResults) {
    return null;
  }

  const curResultIdx = dagResults.findIndex((v) => v.id === dagResultId);
  if (curResultIdx === -1) {
    return null;
  }

  const laterResult = dagResults[curResultIdx - 1];
  const earlierResult = dagResults[curResultIdx + 1];

  return (
    <Box display="flex" width="100%">
      <Tooltip title="Previous Run" arrow>
        <Box sx={{ px: 0, flex: 1 }}>
          <Button
            sx={{ fontSize: '28px', width: '100%' }}
            variant="text"
            onClick={() => {
              navigate(`/workflow/${workflowId}/result/${earlierResult.id}`);
            }}
            disabled={!earlierResult}
          >
            <FontAwesomeIcon icon={faChevronLeft} />
          </Button>
        </Box>
      </Tooltip>

      <Tooltip title="Next Run" arrow>
        <Box sx={{ px: 0, flex: 1 }}>
          <Button
            sx={{ fontSize: '28px', width: '100%' }}
            variant="text"
            onClick={() => {
              navigate(`/workflow/${workflowId}/result/${laterResult.id}`);
            }}
            disabled={!laterResult}
          >
            <FontAwesomeIcon icon={faChevronRight} />
          </Button>
        </Box>
      </Tooltip>
    </Box>
  );
};

export default WorkflowResultNavigator;
