import { faUpRightFromSquare } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Box, Link, Typography } from '@mui/material';

import { UserProfile } from '../..';
import { useDagResultsGetQuery } from '../../handlers/AqueductApi';
import { OperatorsForIntegrationItem } from '../../reducers/integration';
import { Integration } from '../../utils/integrations';
import { ListWorkflowSummary } from '../../utils/workflows';
import { StatusIndicator } from './workflowStatus';

type WorkflowSummaryCardProps = {
  workflow?: ListWorkflowSummary;
  operators: OperatorsForIntegrationItem[];
  integration: Integration;
};

export const WorkflowSummaryCard: React.FC<WorkflowSummaryCardProps> = ({
  workflow,
  operators,
  integration
}) => {

  if (!workflow) {
    return null;
  }

  const workflowLink = `/workflow/${workflow.id}`;

  return (
    <Box
      sx={{
        width: '240px',
        minHeight: '80px',
        backgroundColor: 'gray.100',
        marginBottom: '16px',
        marginRight: '16px',
        borderRadius: '8px',
        py: '8px',
      }}
    >
      <Box sx={{ display: 'flex', flexDirection: 'row', alignItems: 'center' }}>
        <Box
          sx={{
            marginRight: '4px',
            paddingLeft: '16px',
            textOverflow: 'ellipsis',
          }}
        >
          <StatusIndicator status={workflow.status} />
        </Box>
        <Box sx={{ marginRight: '4px' }}>
          <Typography
            variant="body1"
            sx={{
              maxWidth: '150px',
              overflow: 'hidden',
              whiteSpace: 'nowrap',
              textOverflow: 'ellipsis',
              fontSize: '16px',
              my: 0,
            }}
          >
            {workflow.name}
          </Typography>
          <Typography variant="body1" sx={{ fontSize: '8px' }}>
            {new Date(workflow.last_run_at * 1000).toLocaleString()}
          </Typography>
        </Box>
        <Box sx={{ marginLeft: 'auto', marginRight: '16px' }}>
          <Link sx={{ color: 'black' }} target="_blank" href={workflowLink}>
            <FontAwesomeIcon icon={faUpRightFromSquare} />
          </Link>
        </Box>
      </Box>

      <Box sx={{ paddingLeft: '22px' }}>
        <Typography variant="body1" sx={{ fontSize: '12px', my: 0 }}>
          {operators.length} {operators.length > 1 ? 'operators' : 'operator'}{' '}
          using {integration.name}
        </Typography>
      </Box>
    </Box>
  );
};

export default WorkflowSummaryCard;
