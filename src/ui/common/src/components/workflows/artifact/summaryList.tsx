import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Link } from '@mui/material';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React from 'react';
import { Link as RouterLink } from 'react-router-dom';

import { ArtifactResultResponse } from '../../../handlers/responses/artifact';
import { theme } from '../../../styles/theme/theme';
import { getPathPrefix } from '../../../utils/getPathPrefix';
import { artifactTypeToIconMapping } from '../nodes/nodeTypes';

type Props = {
  title: string;
  workflowId: string;
  dagResultId: string;
  artifactResults: ArtifactResultResponse[];
  // When appearance is set to 'value', we will display the value
  // instead of a link whenever possible.
  appearance: 'value' | 'link';
};

const SummaryList: React.FC<Props> = ({
  title,
  workflowId,
  dagResultId,
  artifactResults,
  appearance,
}) => {
  const items = artifactResults.map((artifactResult, index) => {
    let content = null,
      link = null;
    if (artifactResult.result?.content_serialized && appearance === 'value') {
      content = artifactResult.result.content_serialized;
    } else {
      link = `${getPathPrefix()}/workflow/${workflowId}/result/${dagResultId}/artifact/${
        artifactResult.id
      }`;
      content = artifactResult.name;
    }

    const element = (
      <Box
        key={artifactResult.id}
        display="flex"
        p={1}
        sx={{
          alignItems: 'center',
          '&:hover': { backgroundColor: 'gray.100' },
          borderBottom:
            index === artifactResults.length - 1
              ? ''
              : `1px solid ${theme.palette.gray[400]}`,
        }}
      >
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
        <Link
          to={link}
          component={RouterLink as any}
          sx={{ textDecoration: 'none' }}
          key={artifactResult.id}
        >
          {element}
        </Link>
      );
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
