import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Link } from '@mui/material';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React from 'react';
import { Link as RouterLink } from 'react-router-dom';

import { NodesMap } from '../../../handlers/responses/node';
import { theme } from '../../../styles/theme/theme';
import { getPathPrefix } from '../../../utils/getPathPrefix';
import { OperatorType } from '../../../utils/operators';
import { operatorTypeToIconMapping } from '../nodes/Node';

type Props = {
  title: string;
  dagId: string;
  dagResultId?: string;
  workflowId: string;
  nodes: NodesMap;
  operatorIds: string[];
};

const SummaryList: React.FC<Props> = ({
  title,
  workflowId,
  dagId,
  dagResultId,
  nodes,
  operatorIds,
}) => {
  const dagLinkSegment = dagResultId ? `result/${dagResultId}` : `dag/${dagId}`;
  const items = operatorIds.map((id, index) => {
    let link = `${getPathPrefix()}/workflow/${workflowId}/${dagLinkSegment}/operator/${id}`;

    const op = nodes.operators[id];
    const opType = op.spec?.type;
    if (opType === OperatorType.SystemMetric || opType == OperatorType.Metric) {
      link = `${getPathPrefix()}/workflow/${workflowId}/${dagLinkSegment}/metric/${id}`;
    }

    if (opType === OperatorType.Check) {
      link = `${getPathPrefix()}/workflow/${workflowId}/${dagLinkSegment}/check/${id}`;
    }

    return (
      <Link to={link} component={RouterLink} underline="none" key={id}>
        <Box
          display="flex"
          p={1}
          sx={{
            alignItems: 'center',
            '&:hover': { backgroundColor: 'gray.100' },
            borderBottom:
              index === operatorIds.length - 1
                ? ''
                : `1px solid ${theme.palette.gray[400]}`,
          }}
        >
          {!!opType && (
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
                icon={operatorTypeToIconMapping[opType]}
              />
            </Box>
          )}

          <Typography ml="16px">{op.name}</Typography>
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
