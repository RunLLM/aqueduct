import { faUpRightFromSquare } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Box, Link, Typography } from '@mui/material';

import { OperatorsForIntegrationItem } from '../../reducers/integration';
import { ListWorkflowSummary } from '../../utils/workflows';
import { StatusIndicator } from './workflowStatus';
import { Integration } from '../../utils/integrations';

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

  const workflowLink: string = `/workflow/${workflow.id}`;

  return (
    <Box
      sx={{
        width: '240px',
        minHeight: '80px',
        backgroundColor: 'gray.100',
        marginBottom: '16px',
        borderRadius: '8px',
      }}
    >
      <Box sx={{ display: 'flex', flexDirection: 'row', alignItems: 'center' }}>
        <Box
          sx={{
            marginRight: '4px',
            paddingLeft: '16px',
            maxWidth: '150px',
            textOverflow: 'ellipsis',
          }}
        >
          <StatusIndicator status={workflow.status} />
        </Box>
        <Box sx={{ marginRight: '4px' }}>
          <Typography variant="body1" sx={{ fontSize: '16px', my: 0 }}>
            {workflow.name}
          </Typography>
          <Typography variant="body1" sx={{ fontSize: '8px' }}>
            {new Date(workflow.last_run_at * 1000).toLocaleString()}
          </Typography>
        </Box>
        <Box sx={{ marginLeft: '16px', color: 'black' }}>
          <Link sx={{ color: 'black' }} href={workflowLink}>
            <FontAwesomeIcon icon={faUpRightFromSquare} />
          </Link>
        </Box>
      </Box>

      <Box sx={{ paddingLeft: '22px' }}>
          <Typography variant="body1" sx={{ fontSize: '12px', my: 0 }}>
            {operators.length} {operators.length > 1 ? 'operators' : 'operator'} using {integration.name}
          </Typography>
        </Box>
    </Box>
  );
};

export default WorkflowSummaryCard;
