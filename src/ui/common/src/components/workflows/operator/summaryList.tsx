import { faChevronRight } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Link, List, ListItem } from '@mui/material';
import Accordion from '@mui/material/Accordion';
import AccordionDetails from '@mui/material/AccordionDetails';
import AccordionSummary from '@mui/material/AccordionSummary';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React, { useState } from 'react';
import { Link as RouterLink } from 'react-router-dom';

import { OperatorResultResponse } from '../../../handlers/responses/operator';
import { getPathPrefix } from '../../../utils/getPathPrefix';
import { OperatorType } from '../../../utils/operators';
import { operatorTypeToIconMapping } from '../nodes/nodeTypes';

type Props = {
  title: string;
  workflowId: string;
  dagResultId: string;
  operatorResults: OperatorResultResponse[];
  initiallyExpanded: boolean;
};

const listStyle = {
  width: '100%',
  maxWidth: 360,
  bgcolor: 'background.paper',
};

const SummaryList: React.FC<Props> = ({
  title,
  workflowId,
  dagResultId,
  operatorResults,
  initiallyExpanded,
}) => {
  const [expanded, setExpanded] = useState(initiallyExpanded);
  const items = operatorResults.map((opResult) => {
    let link = `${getPathPrefix()}/workflow/${workflowId}/result/${dagResultId}/operator/${
      opResult.id
    }`;
    const opType = opResult.spec?.type;
    if (opType === OperatorType.SystemMetric || opType == OperatorType.Metric) {
      link = `${getPathPrefix()}/workflow/${workflowId}/result/${dagResultId}/metric/${
        opResult.id
      }`;
    }

    if (opType === OperatorType.Check) {
      link = `${getPathPrefix()}/workflow/${workflowId}/result/${dagResultId}/check/${
        opResult.id
      }`;
    }

    return (
      <ListItem divider key={opResult.id}>
        <Box display="flex">
          {!!opResult.spec?.type && (
            <Box
              sx={{
                width: '16px',
                height: '16px',
                color: 'rgba(0,0,0,0.54)',
              }}
            >
              <FontAwesomeIcon
                icon={operatorTypeToIconMapping[opResult.spec.type]}
              />
            </Box>
          )}
          <Link
            to={link}
            component={RouterLink as any}
            sx={{ marginLeft: '16px' }}
            underline="none"
          >
            {opResult.name}
          </Link>
        </Box>
      </ListItem>
    );
  });

  return (
    <Accordion
      expanded={expanded}
      onChange={() => {
        setExpanded(!expanded);
      }}
    >
      <AccordionSummary
        expandIcon={<FontAwesomeIcon icon={faChevronRight} />}
        sx={{
          '& .MuiAccordionSummary-expandIconWrapper.Mui-expanded': {
            transform: 'rotate(90deg)',
          },
        }}
        aria-controls="input-accordion-content"
        id="input-accordion-header"
      >
        <Typography
          sx={{ width: '33%', flexShrink: 0 }}
          variant="h5"
          component="div"
          marginBottom="8px"
        >
          {title}
        </Typography>
      </AccordionSummary>
      <AccordionDetails>
        <List sx={listStyle}>{items}</List>
      </AccordionDetails>
    </Accordion>
  );
};

export default SummaryList;
