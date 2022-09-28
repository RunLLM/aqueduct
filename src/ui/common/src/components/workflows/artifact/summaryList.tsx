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

import { ArtifactResultResponse } from '../../../handlers/responses/artifact';
import { getPathPrefix } from '../../../utils/getPathPrefix';
import { artifactTypeToIconMapping } from '../nodes/nodeTypes';

type Props = {
  title: string;
  workflowId: string;
  dagResultId: string;
  artifactResults: ArtifactResultResponse[];
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
  artifactResults,
  initiallyExpanded,
}) => {
  const [expanded, setExpanded] = useState(initiallyExpanded);
  const items = artifactResults.map((artifactResult) => {
    let content = null;
    if (artifactResult.result?.content_serialized) {
      content = (
        <Typography variant="body1">
          {artifactResult.result.content_serialized}
        </Typography>
      );
    } else {
      content = (
        <Box display="flex">
          <Box
            sx={{
              width: '16px',
              height: '16px',
              color: 'rgba(0,0,0,0.54)',
            }}
          >
            <FontAwesomeIcon
              icon={artifactTypeToIconMapping[artifactResult.type]}
            />
          </Box>
          <Link
            to={`${getPathPrefix()}/workflow/${workflowId}/result/${dagResultId}/artifact/${
              artifactResult.id
            }`}
            component={RouterLink as any}
            sx={{ marginLeft: '16px' }}
            underline="none"
          >
            {artifactResult.name}
          </Link>
        </Box>
      );
    }
    return (
      <ListItem divider key={artifactResult.id}>
        {content}
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
