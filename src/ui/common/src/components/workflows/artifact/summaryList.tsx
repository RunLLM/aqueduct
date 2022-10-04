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
import { theme } from '../../../styles/theme/theme';

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
  const items = artifactResults.map((artifactResult, index) => {
    let content = null, link = null;
    if (artifactResult.result?.content_serialized) {
      content = artifactResult.result.content_serialized;
    } else {
      link = `${getPathPrefix()}/workflow/${workflowId}/result/${dagResultId}/artifact/${artifactResult.id}}`;
      content = artifactResult.name;
    }
    
    let element =  (
      <Box display="flex" p={1} sx={{ alignItems: 'center', '&:hover': { backgroundColor: 'gray.100' }, borderBottom: index === artifactResults.length  - 1 ? '' : `1px solid ${theme.palette.gray[400]}` }}>
        <Box display="flex" sx={{ alignItems: 'center' }}>
          <Box
            sx={{
              width: '16px',
              height: '16px',
            }}
          >
            <FontAwesomeIcon
              fontSize="16px"
              color={`${theme.palette.gray[700]}`}
              icon={artifactTypeToIconMapping[artifactResult.type]}
            />
          </Box>
          <Typography ml="16px">{content}</Typography>
        </Box>
      </Box>
    );

    if (link) {
      return (
        <Link to={link} component={RouterLink as any} sx={{ textDecoration: 'none' }}>
          {element}
        </Link>
      )
    }

    return element;
  });

  return (
    <Box>
      <Typography variant="h6" mb="8px" fontWeight="normal">
        {title}
      </Typography>
      {items}
    </Box>
  );
};

export default SummaryList;
