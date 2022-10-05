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
};

const SummaryList: React.FC<Props> = ({
  title,
  workflowId,
  dagResultId,
  artifactResults,
}) => {
  const items = artifactResults.map((artifactResult, idx) => {
    return (
      <Link
        key={artifactResult.id}
        to={`${getPathPrefix()}/workflow/${workflowId}/result/${dagResultId}/artifact/${
          artifactResult.id
        }`}
        component={RouterLink as any}
        underline="none"
      >
        <Box
          display="flex"
          p={1}
          sx={{
            alignItems: 'center',
            '&:hover': { backgroundColor: 'gray.100' },
            borderBottom:
              idx === artifactResults.length - 1
                ? ''
                : `1px solid ${theme.palette.gray[400]}`,
          }}
        >
          <Box
            width="16px"
            height="16px"
            alignItems="center"
            display="flex"
            flexDirection="column"
          >
            <FontAwesomeIcon
              fontSize="16px"
              color={`${theme.palette.gray[700]}`}
              icon={artifactTypeToIconMapping[artifactResult.type]}
            />
          </Box>
          <Typography ml="16px">{artifactResult.name}</Typography>
        </Box>
      </Link>
    );
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
