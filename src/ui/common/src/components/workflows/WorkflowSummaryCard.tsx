import { faUpRightFromSquare } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Box, Link, Typography } from '@mui/material';
import React from 'react';

import { OperatorsForIntegrationItem } from '../../reducers/integration';
import { Integration } from '../../utils/integrations';
import { ListWorkflowSummary } from '../../utils/workflows';
import { TruncatedText } from '../integrations/cards/text';
import { StatusIndicator } from './workflowStatus';

export type WorkflowSummaryCardProps = {
  workflow?: ListWorkflowSummary;
  operators: OperatorsForIntegrationItem[];
  integration: Integration;
};

export const WorkflowSummaryCard: React.FC<WorkflowSummaryCardProps> = ({
  workflow,
  operators,
  integration,
}) => {
  if (!workflow) {
    return null;
  }

  const workflowLink = `/workflow/${workflow.id}`;

  return (
    <Box
      sx={{
        width: '325px',
        minHeight: '96px',
        backgroundColor: '#F8F8F8',
        marginBottom: '16px',
        marginRight: '16px',
        borderRadius: '8px',
        py: '8px',
      }}
    >
      <Box sx={{ display: 'flex', flexDirection: 'row', alignItems: 'center' }}>
        <Box>
          <Box sx={{ display: 'flex', marginLeft: '8px' }}>
            <StatusIndicator status={workflow.status} size="16px" />
            <Typography
              variant="body1"
              sx={{
                marginLeft: '8px',
                maxWidth: '265px',
                overflow: 'hidden',
                whiteSpace: 'nowrap',
                textOverflow: 'ellipsis',
                fontSize: '16px',
                my: 0,
              }}
            >
              {workflow.name}
            </Typography>
          </Box>
          <Typography
            variant="body1"
            sx={{
              marginLeft: '8px',
              fontWeight: 400,
              fontSize: '10px',
              color: '#858585',
            }}
          >
            {workflow.last_run_at
              ? new Date(workflow.last_run_at * 1000).toLocaleString()
              : `N/A`}
          </Typography>
        </Box>
        <Box
          sx={{
            marginLeft: 'auto',
            marginRight: '8px',
            alignSelf: 'flex-start',
          }}
        >
          <Link sx={{ color: 'black' }} target="_blank" href={workflowLink}>
            <FontAwesomeIcon icon={faUpRightFromSquare} />
          </Link>
        </Box>
      </Box>

      <Box
        sx={{
          position: 'relative',
          top: '20px',
          left: '8px',
          textAlign: 'left',
        }}
      >
        <TruncatedText variant="caption" sx={{ fontWeight: 300 }}>
          {operators.length} {operators.length > 1 ? 'operators' : 'operator'}{' '}
          using {integration.name}
        </TruncatedText>
      </Box>
    </Box>
  );
};

export default WorkflowSummaryCard;
