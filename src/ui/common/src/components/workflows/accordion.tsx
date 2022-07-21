import { Link } from '@mui/material';
import MuiAccordion from '@mui/material/Accordion';
import MuiAccordionDetails from '@mui/material/AccordionDetails';
import MuiAccordionSummary from '@mui/material/AccordionSummary';
import Typography from '@mui/material/Typography';
import React from 'react';

import { OperatorsForIntegrationItem } from '../../reducers/integrationOperators';
import { getPathPrefix } from '../../utils/getPathPrefix';
import { ListWorkflowSummary } from '../../utils/workflows';
import OperatorsTable from '../tables/operatorsTable';

type Props = {
  workflow?: ListWorkflowSummary;
  operators: OperatorsForIntegrationItem[];
};

const WorkflowAccordion: React.FC<Props> = ({ workflow, operators }) => {
  return (
    <MuiAccordion>
      <MuiAccordionSummary aria-controls="panel1d-content" id="panel1d-header">
        {workflow ? (
          <Link
            underline="none"
            color="inherit"
            href={`${getPathPrefix()}/workflow/${workflow.id}`}
          >
            <Typography variant="body1"> {workflow.name} </Typography>
          </Link>
        ) : (
          <Typography variant="body1"> Unknown workflow </Typography>
        )}
      </MuiAccordionSummary>
      <MuiAccordionDetails>
        <OperatorsTable operators={operators} />
      </MuiAccordionDetails>
    </MuiAccordion>
  );
};

export default WorkflowAccordion;
