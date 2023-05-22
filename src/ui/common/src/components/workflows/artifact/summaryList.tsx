import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Link } from '@mui/material';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React from 'react';
import { Link as RouterLink } from 'react-router-dom';

import { NodeResultsMap, NodesMap } from '../../../handlers/responses/node';
import { theme } from '../../../styles/theme/theme';
import { getPathPrefix } from '../../../utils/getPathPrefix';
import { OperatorType } from '../../../utils/operators';
import { artifactTypeToIconMapping } from '../nodes/Node';

type Props = {
  title: string;
  dagId: string;
  dagResultId?: string;
  workflowId: string;
  nodes: NodesMap;
  nodeResults?: NodeResultsMap;
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
  nodeResults,
  artifactIds,
  appearance = 'link',
  collapsePrimitives = true,
}) => {
  const items = artifactIds
    .map((artfId, index) => {
      const artf = nodes.artifacts[artfId];
      const artfResult = (nodeResults?.artifacts ?? {})[artfId];
      if (!artf) {
        return null;
      }

      let content = null,
        link = null;

      let linkType = 'artifact';
      let linkTarget = artfId;
      const dagLinkSegment = dagResultId
        ? `result/${dagResultId}`
        : `dag/${dagId}`;

      const fromOpType = nodes.operators[artf.input]?.spec?.type;
      if (
        fromOpType === OperatorType.Metric ||
        fromOpType === OperatorType.Check
      ) {
        // For checks & metrics, we want to the URL to be of the form /metric/{operatorId}, which is why we set both the
        // linkType and linkTarget here.
        linkType = fromOpType;
        linkTarget = artf.input;
      }

      if (
        artfResult?.content_serialized &&
        appearance === 'value' &&
        collapsePrimitives
      ) {
        content = artfResult.content_serialized;
      } else if (artfResult?.content_serialized) {
        // Show the name and the value and link it.
        link = `${getPathPrefix()}/workflow/${workflowId}/${dagLinkSegment}/${linkType}/${linkTarget}`;
        content = `${artf.name} (${artfResult.content_serialized})`;
      } else {
        // Show only the name and link it.
        link = `${getPathPrefix()}/workflow/${workflowId}/${dagLinkSegment}/${linkType}/${linkTarget}`;
        content = artf.name;
      }

      const element = (
        <Box
          key={artfId}
          display="flex"
          p={1}
          sx={{
            alignItems: 'center',
            '&:hover': { backgroundColor: 'gray.100' },
            borderBottom:
              index === artifactIds.length - 1
                ? ''
                : `1px solid ${theme.palette.gray[400]}`,
          }}
        >
          <Box display="flex" sx={{ alignItems: 'center' }}>
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
                icon={artifactTypeToIconMapping[artf.type]}
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
            key={artfId}
          >
            {element}
          </Link>
        );
      }

      return element;
    })
    .filter((x) => !!x);

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
