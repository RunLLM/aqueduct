import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React from 'react';

import {
  GoogleSheetsExtractParams,
  GoogleSheetsLoadParams,
  MongoDBExtractParams,
  Operator,
  OperatorType,
  RelationalDBExtractParams,
  RelationalDBLoadParams,
} from '../../utils/operators';
import { CodeBlock } from '../CodeBlock';

type Props = {
  operator: Operator;
  textColor?: string;
};

const OperatorParametersOverview: React.FC<Props> = ({
  operator,
  textColor,
}) => {
  if (operator.spec.type === OperatorType.Extract) {
    const exParams = operator.spec.extract?.parameters;
    if (!exParams) {
      return null;
    }

    // These checks tries to distinguish googlesheet vs relational
    // extracts based on the fields of type union exParams.
    if ('query' in exParams || 'queries' in exParams) {
      const relationalParams = exParams as RelationalDBExtractParams;
      if (!!relationalParams.queries) {
        return (
          <Typography variant="body2" color={textColor}>
            {`${relationalParams.queries.length} chained queries.`}
          </Typography>
        );
      }

      return (
        <Typography variant="body2" color={textColor}>
          <strong>query: </strong>
          <code>{relationalParams.query}</code>
        </Typography>
      );
    } else if ('spreadsheet_id' in exParams) {
      return (
        <Typography variant="body2" color={textColor}>
          <strong>spreadsheet ID: </strong>
          {(exParams as GoogleSheetsExtractParams).spreadsheet_id}
        </Typography>
      );
    } else if ('query_serialized' in exParams) {
      const mongoDbParams = exParams as MongoDBExtractParams;
      return (
        <Box>
          <Typography variant="body2" color={textColor}>
            <strong>collection: </strong>
            <code>{mongoDbParams.collection}</code>
          </Typography>
          <Typography variant="body2" color={textColor} mb={1}>
            <strong>query: </strong>
          </Typography>
          <CodeBlock language="json">
            {JSON.stringify(
              // pretty print
              JSON.parse(mongoDbParams.query_serialized),
              null,
              2
            )}
          </CodeBlock>
        </Box>
      );
    }
  } else if (operator.spec.type === OperatorType.Load) {
    const loadParams = operator.spec.load?.parameters;
    if (!loadParams) {
      return null;
    }

    // These checks tries to distinguish googlesheet vs relational
    // loads based on the fields of type union laodParams.
    if ('table' in loadParams) {
      return (
        <Box>
          <Typography variant="body2" color={textColor}>
            <strong>table: </strong>
            {(loadParams as RelationalDBLoadParams).table}
          </Typography>
          <Typography
            variant="body2"
            color={textColor}
            sx={{ marginTop: '2px' }}
          >
            <strong>update_mode: </strong>
            {(loadParams as RelationalDBLoadParams).update_mode}
          </Typography>
        </Box>
      );
    } else if ('filepath' in loadParams) {
      return (
        <Box>
          <Typography variant="body2" color={textColor}>
            <strong>filepath: </strong>
            {(loadParams as GoogleSheetsLoadParams).filepath}
          </Typography>
          <Typography
            variant="body2"
            color={textColor}
            sx={{ marginTop: '2px' }}
          >
            <strong>save_mode: </strong>
            {(loadParams as GoogleSheetsLoadParams).save_mode}
          </Typography>
        </Box>
      );
    }
  }

  return null;
};

export default OperatorParametersOverview;
