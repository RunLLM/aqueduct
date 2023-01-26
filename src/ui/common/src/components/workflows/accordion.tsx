import { faChevronRight } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import MuiAccordion, { AccordionProps } from '@mui/material/Accordion';
import AccordionDetails from '@mui/material/AccordionDetails';
import MuiAccordionSummary, {
  AccordionSummaryProps,
} from '@mui/material/AccordionSummary';
//import styled from '@emotion/styled';
import { styled } from '@mui/material/styles';
import Typography from '@mui/material/Typography';
import React from 'react';

import { OperatorsForIntegrationItem } from '../../reducers/integration';
import { ListWorkflowSummary } from '../../utils/workflows';
import OperatorsTable from '../tables/operatorsTable';

type Props = {
  expanded: boolean;
  handleExpand?: () => void;
  workflow?: ListWorkflowSummary;
  operators: OperatorsForIntegrationItem[];
};

const Accordion = styled((props: AccordionProps) => (
  <MuiAccordion disableGutters elevation={0} square {...props} />
))(({ theme }) => ({
  border: `1px solid ${theme.palette.divider}`,
  '&:not(:last-child)': {
    borderBottom: 0,
  },
  '&:before': {
    display: 'none',
  },
}));

const AccordionSummary = styled((props: AccordionSummaryProps) => (
  <MuiAccordionSummary
    expandIcon={<FontAwesomeIcon icon={faChevronRight} />}
    {...props}
  />
))(({ theme }) => ({
  backgroundColor:
    theme.palette.mode === 'dark'
      ? 'rgba(255, 255, 255, .05)'
      : 'rgba(0, 0, 0, .03)',
  '& .MuiAccordionSummary-expandIconWrapper.Mui-expanded': {
    transform: 'rotate(90deg)',
  },
}));

const WorkflowAccordion: React.FC<Props> = ({
  expanded,
  handleExpand,
  workflow,
  operators,
}) => {
  return (
    <Accordion expanded={expanded} onChange={handleExpand}>
      <AccordionSummary aria-controls="panel1d-content" id="panel1d-header">
        {workflow ? (
          <Typography variant="body1"> {workflow.name} </Typography>
        ) : (
          <Typography variant="body1"> Unknown workflow </Typography>
        )}
      </AccordionSummary>
      <AccordionDetails>
        <OperatorsTable workflow={workflow} operators={operators} />
      </AccordionDetails>
    </Accordion>
  );
};

export default WorkflowAccordion;
