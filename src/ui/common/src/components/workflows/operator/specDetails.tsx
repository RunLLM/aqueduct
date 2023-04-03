import { faCircleDown } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import { BlobReader, TextWriter, ZipReader } from '@zip.js/zip.js';
import React, { useEffect, useState } from 'react';

import MultiFileViewer from '../../../components/MultiFileViewer';
import { InfoTooltip } from '../../../components/pages/components/InfoTooltip';
import { OperatorResultResponse } from '../../../handlers/responses/operatorDeprecated';
import UserProfile from '../../../utils/auth';
import {
  exportFunction,
  GoogleSheetsExtractParams,
  handleExportFunction,
  hasFile,
  MongoDBExtractParams,
  OperatorType,
  PREV_TABLE_TAG,
  RelationalDBExtractParams,
} from '../../../utils/operators';
import { CodeBlock } from '../../CodeBlock';
import { Button } from '../../primitives/Button.styles';

const MAX_FILE_RENDER_SIZE = 100000000; // 100M

type Props = {
  user: UserProfile;
  operator: OperatorResultResponse;
};

// Checked with file size=313285391 and handles that smoothly once loaded. However, takes a while to load.
const SpecDetails: React.FC<Props> = ({ user, operator }) => {
  const [files, setFiles] = useState({
    '': {
      path: '',
      language: 'plaintext',
      content: '',
    },
  });

  const setFileHelper = (prevState, file, fileContents) => {
    const nextState = { ...prevState };
    const pathList = file.filename.split('/');
    let base = nextState;
    pathList.forEach((section, i) => {
      // Create a key for each first-level subfolder
      if (!Object.keys(base).includes(section)) {
        base[section] = {};
      }
      if (!file.directory && i + 1 === pathList.length) {
        // Include the file metadata
        base[section] = fileContents;
      } else {
        // Go into the subfolder
        base = base[section];
      }
    });
    return nextState;
  };

  useEffect(() => {
    async function getFilesBlob() {
      // This is the function used to retrieve the contents in the function that generates the operator's zip file.
      const blob = await exportFunction(user, operator.id);
      if (blob) {
        const reader = new ZipReader(new BlobReader(blob));
        const entries = await reader.getEntries();
        entries.forEach((file) => {
          let language = 'plaintext';
          if (file.filename.endsWith('.py')) {
            language = 'python';
          }
          if (file.uncompressedSize < MAX_FILE_RENDER_SIZE) {
            file.getData(new TextWriter()).then((content) => {
              setFiles((prevState) =>
                setFileHelper(prevState, file, {
                  path: file.filename,
                  language: language,
                  content: content,
                })
              );
            });
          } else {
            setFiles((prevState) =>
              setFileHelper(prevState, file, {
                path: file.filename,
                language: 'plaintext',
                content:
                  'We do not support viewing such large files.\nPlease download this file instead.',
              })
            );
          }
        });
        await reader.close();
      }
    }

    if (hasFile(operator?.spec?.type)) {
      getFilesBlob();
    }
  }, [operator, user]);

  if (operator.spec === undefined) {
    return null;
  }

  if (hasFile(operator?.spec?.type)) {
    return (
      <Box>
        <Box mb={1} display="flex" flexDirection="row">
          <Typography variant="h6" fontWeight="normal">
            Code Preview
          </Typography>
          <Box flexGrow={1} />
          <Button
            onClick={async () => {
              await handleExportFunction(
                user,
                operator.id,
                `${operator.name}.zip`
              );
            }}
            color="secondary"
          >
            <FontAwesomeIcon icon={faCircleDown} />
            <Typography sx={{ ml: 1 }}>{`${operator.name}.zip`}</Typography>
          </Button>
        </Box>
        <MultiFileViewer files={files} defaultFile={operator.name || ''} />
      </Box>
    );
  }

  if (operator?.spec?.type === OperatorType.Extract) {
    const extractParams = operator.spec.extract.parameters;
    let content = null;

    if ('query' in extractParams || 'queries' in extractParams) {
      // relational
      const relationalParams = extractParams as RelationalDBExtractParams;
      const renderQuery = (q: string) => (
        <CodeBlock language="sql">{q}</CodeBlock>
      );
      let tooltips = '';
      const chainTagTooltips =
        '`$` refers to the output of the previous query.';

      if (!!relationalParams.queries && relationalParams.queries.length > 0) {
        const queries = relationalParams.queries;
        content = (
          <Box display="flex" flexDirection="column">
            {relationalParams.queries.map((q, idx) => (
              <Box mb={1} key={`extract-query-${idx}`}>
                {renderQuery(q)}
              </Box>
            ))}
          </Box>
        );
        const hasChainTag = queries.some((q) => q.includes(PREV_TABLE_TAG));
        tooltips = `These queries are chained. ${
          hasChainTag ? chainTagTooltips : ''
        }`;
      } else {
        content = renderQuery(relationalParams.query);
      }

      return (
        <Box>
          <Box display="flex" flexDirection="row" marginBottom={1}>
            <Typography variant="h6" fontWeight="normal" alignContent="center">
              Query Details
            </Typography>
            {tooltips && (
              <InfoTooltip tooltipText={tooltips} placement="right" />
            )}
          </Box>
          {content}
        </Box>
      );
    }

    // MongoDB
    if ('query_serialized' in extractParams) {
      const params = extractParams as MongoDBExtractParams;
      return (
        <Box>
          <Box display="flex" flexDirection="row" marginBottom={1}>
            <Typography variant="h6" fontWeight="normal" alignContent="center">
              Query Details
            </Typography>
          </Box>
          <Typography variant="body2" alignContent="center" marginBottom={1}>
            <strong>Collection: </strong>
            {params.collection}
          </Typography>
          <Typography variant="body2" alignContent="center" marginBottom={1}>
            <strong>Query: </strong>
          </Typography>
          <CodeBlock language="json">
            {JSON.stringify(
              // pretty print
              JSON.parse(params.query_serialized),
              null,
              2
            )}
          </CodeBlock>
        </Box>
      );
    }

    // google sheet
    const googleSheetsParams = extractParams as GoogleSheetsExtractParams;
    return (
      <Box>
        <Typography variant="h6" fontWeight="normal" mb={1}>
          Spreadsheet Details
        </Typography>
        <Typography variant="body1" mb={1}>
          <strong>Spreadsheet ID: </strong>
          {googleSheetsParams.spreadsheet_id}
        </Typography>
        <Typography variant="body1">
          <strong>Query: </strong>
          {googleSheetsParams.query}
        </Typography>
      </Box>
    );
  }

  return null;
};

export default SpecDetails;
