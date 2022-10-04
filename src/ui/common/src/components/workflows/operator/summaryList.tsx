import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Link } from '@mui/material';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React, { useState } from 'react';
import { Link as RouterLink } from 'react-router-dom';

import { OperatorResultResponse } from '../../../handlers/responses/operator';
import { theme } from '../../../styles/theme/theme';
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
  const items = operatorResults.map((opResult, index) => {
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
      <Link to={link} component={RouterLink as any} underline="none">
        <Box
          display="flex"
          p={1}
          sx={{
            alignItems: 'center',
            '&:hover': { backgroundColor: 'gray.100' },
            borderBottom:
              index === operatorResults.length - 1
                ? ''
                : `1px solid ${theme.palette.gray[400]}`,
          }}
        >
          {!!opResult.spec?.type && (
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
                icon={operatorTypeToIconMapping[opResult.spec.type]}
              />
            </Box>
          )}

          <Typography ml="16px">{opResult.name}</Typography>
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
