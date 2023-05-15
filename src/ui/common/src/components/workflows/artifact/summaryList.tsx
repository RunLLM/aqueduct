import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Link } from '@mui/material';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React from 'react';
import { Link as RouterLink } from 'react-router-dom';

import { theme } from '../../../styles/theme/theme';
import { getPathPrefix } from '../../../utils/getPathPrefix';
import { OperatorType } from '../../../utils/operators';
import { artifactTypeToIconMapping } from '../nodes/Node';
import { NodeResultsMap, NodesMap } from '../../../handlers/responses/node';

type Props = {
  title: string;
  dagId: string;
  dagResultId?: string;
  workflowId: string;
  nodes: NodesMap;
  nodeResults: NodeResultsMap;
  artifactIds: string[];
  collapsePrimitives?: boolean;
  // When appearance is set to 'value', we will display the value
  // instead of a link whenever possible.
  appearance?: 'value' | 'link';
};

const SummaryList: React.FC<Props> = ({
  title,
  workflowId,
  dagId,
  dagResultId,
  nodes,
  artifactIds,
  appearance = 'link',
  collapsePrimitives = true,
}) => {
  const items = artifactIds.map((artfId, index) => {
    const artf = nodes.artifacts[artfId]
    const artfResult = nodeR
    let content = null,
      link = null;

    let linkType = 'artifact';
    let linkTarget = artfId;
    const dagLinkSegment = dagResultId
      ? `result/${dagResultId}`
      : `dag/${dagId}`;

    const fromOpType = nodes.operators[artf.artifact.input]?.spec?.type
    if (
      fromOpType === OperatorType.Metric ||
      fromOpType === OperatorType.Check
    ) {
      // For checks & metrics, we want to the URL to be of the form /metric/{operatorId}, which is why we set both the
      // linkType and linkTarget here.
      linkType = fromOpType;
      linkTarget = artf.artifact.input;
    }

    if (
      artf.result?.content_serialized &&
      appearance === 'value' &&
      collapsePrimitives
    ) {
      content = artf.result.content_serialized;
    } else if (artf.result?.content_serialized) {
      // Show the name and the value and link it.
      link = `${getPathPrefix()}/workflow/${workflowId}/${dagLinkSegment}/${linkType}/${linkTarget}`;
      content = `${artf.artifact.name} (${artf.result.content_serialized})`;
    } else {
      // Show only the name and link it.
      link = `${getPathPrefix()}/workflow/${workflowId}/${dagLinkSegment}/${linkType}/${linkTarget}`;
      content = artf.artifact.name;
    }

    const element = (
      <Box
        key={artf.artifact.id}
        display="flex"
        p={1}
        sx={{
          alignItems: 'center',
          '&:hover': { backgroundColor: 'gray.100' },
          borderBottom:
            index === artifacts.length - 1
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
              icon={artifactTypeToIconMapping[artf.artifact.type]}
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
          component={RouterLink}
          sx={{ textDecoration: 'none' }}
          key={artf.artifact.id}
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
