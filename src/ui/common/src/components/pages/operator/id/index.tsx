import {
  faCircleDown,
  faQuestionCircle,
} from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { CircularProgress, Divider } from '@mui/material';
import Box from '@mui/material/Box';
import IconButton from '@mui/material/IconButton';
import Tooltip from '@mui/material/Tooltip';
import Typography from '@mui/material/Typography';
import { BlobReader, TextWriter, ZipReader } from '@zip.js/zip.js';
import React, { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useLocation, useNavigate, useParams } from 'react-router-dom';

import DefaultLayout, {
  SidesheetContentWidth,
} from '../../../../components/layouts/default';
import { BreadcrumbLink } from '../../../../components/layouts/NavBar';
import LogViewer from '../../../../components/LogViewer';
import MultiFileViewer from '../../../../components/MultiFileViewer';
import { handleGetWorkflowDagResult } from '../../../../handlers/getWorkflowDagResult';
import { AppDispatch, RootState } from '../../../../stores/store';
import UserProfile from '../../../../utils/auth';
import { getPathPrefix } from '../../../../utils/getPathPrefix';
import {
  exportFunction,
  GoogleSheetsExtractParams,
  handleExportFunction,
  hasFile,
  LEFT_PARAMS_TAG,
  OperatorType,
  PREV_TABLE_TAG,
  RelationalDBExtractParams,
} from '../../../../utils/operators';
import {
  isInitial,
  isLoading,
  LoadingStatusEnum,
} from '../../../../utils/shared';
import { CodeBlock } from '../../../CodeBlock';
import { Button } from '../../../primitives/Button.styles';
import ArtifactSummaryList from '../../../workflows/artifact/summaryList';
import DetailsPageHeader from '../../components/DetailsPageHeader';
import { LayoutProps } from '../../types';

type OperatorDetailsPageProps = {
  user: UserProfile;
  Layout?: React.FC<LayoutProps>;
  maxRenderSize?: number;
  workflowIdProp?: string;
  workflowDagResultIdProp?: string;
  operatorIdProp?: string;
  sideSheetMode?: boolean;
};

// Checked with file size=313285391 and handles that smoothly once loaded. However, takes a while to load.
const OperatorDetailsPage: React.FC<OperatorDetailsPageProps> = ({
  user,
  Layout = DefaultLayout,
  maxRenderSize = 100000000,
  workflowIdProp,
  workflowDagResultIdProp,
  operatorIdProp,
  sideSheetMode = false,
}) => {
  const dispatch: AppDispatch = useDispatch();
  const navigate = useNavigate();
  let { workflowId, workflowDagResultId, operatorId } = useParams();
  const path = useLocation().pathname;

  if (workflowIdProp) {
    workflowId = workflowIdProp;
  }

  if (workflowDagResultIdProp) {
    workflowDagResultId = workflowDagResultIdProp;
  }

  if (operatorIdProp) {
    operatorId = operatorIdProp;
  }

  const [files, setFiles] = useState({
    '': {
      path: '',
      language: 'plaintext',
      content: '',
    },
  });

  const workflowDagResultWithLoadingStatus = useSelector(
    (state: RootState) =>
      state.workflowDagResultsReducer.results[workflowDagResultId]
  );

  const operator = (workflowDagResultWithLoadingStatus?.result?.operators ??
    {})[operatorId];

  const pathPrefix = getPathPrefix();
  const workflowLink = `${pathPrefix}/workflow/${workflowId}?workflowDagResultId=${workflowDagResultId}`;
  const breadcrumbs = [
    BreadcrumbLink.HOME,
    BreadcrumbLink.WORKFLOWS,
    new BreadcrumbLink(
      workflowLink,
      workflowDagResultWithLoadingStatus?.result?.name ?? 'Workflow'
    ),
    new BreadcrumbLink(path, operator?.name || 'Operator'),
  ];

  useEffect(() => {
    document.title = 'Operator Details | Aqueduct';

    if (
      // Load workflow dag result if it's not cached
      !workflowDagResultWithLoadingStatus ||
      isInitial(workflowDagResultWithLoadingStatus.status)
    ) {
      dispatch(
        handleGetWorkflowDagResult({
          apiKey: user.apiKey,
          workflowId,
          workflowDagResultId,
        })
      );
    }
  }, []);

  useEffect(() => {
    if (!!operator && !sideSheetMode) {
      document.title = `${operator?.name || 'Operator'} | Aqueduct`;
    }
  }, [operator]);

  const logs = operator?.result?.exec_state?.user_logs ?? {};
  const operatorError = operator?.result?.exec_state?.error;

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
      const blob = await exportFunction(user, operatorId);
      if (blob) {
        const reader = new ZipReader(new BlobReader(blob));
        const entries = await reader.getEntries();
        entries.forEach((file) => {
          let language = 'plaintext';
          if (file.filename.endsWith('.py')) {
            language = 'python';
          }
          if (file.uncompressedSize < maxRenderSize) {
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
  }, [operator]);

  if (
    !workflowDagResultWithLoadingStatus ||
    isInitial(workflowDagResultWithLoadingStatus.status) ||
    isLoading(workflowDagResultWithLoadingStatus.status)
  ) {
    return (
      <Layout breadcrumbs={breadcrumbs} user={user}>
        <CircularProgress />
      </Layout>
    );
  }

  // This workflow doesn't exist.
  if (
    workflowDagResultWithLoadingStatus.status.loading ===
    LoadingStatusEnum.Failed
  ) {
    navigate('/404');
    return null;
  }

  const mapArtifacts = (artfIds: string[]) =>
    artfIds
      .map(
        (artifactId) =>
          (workflowDagResultWithLoadingStatus.result?.artifacts ?? {})[
            artifactId
          ]
      )
      .filter((artf) => !!artf);
  const inputs = mapArtifacts(operator.inputs);
  const outputs = mapArtifacts(operator.outputs);

  const operatorPreview = () => {
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
        const parameterTooltips =
          'Contents in `{{}}` refers to the name of parameters.';
        const chainTagTooltips =
          '`$` refers to the outputs of the previous query.';

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
          const hasParameters = queries.some((q) =>
            q.includes(LEFT_PARAMS_TAG)
          );
          tooltips = `This is a chained query. ${
            hasChainTag ? chainTagTooltips : ''
          } ${hasParameters ? parameterTooltips : ''}`;
        } else {
          content = renderQuery(relationalParams.query);
          if (relationalParams.query.includes(LEFT_PARAMS_TAG)) {
            tooltips = parameterTooltips;
          }
        }

        return (
          <Box>
            <Box display="flex" flexDirection="row" marginBottom={1}>
              <Typography
                variant="h6"
                fontWeight="normal"
                alignContent="center"
              >
                Query Details
              </Typography>
              {tooltips && (
                <Tooltip arrow placement="right" title={tooltips}>
                  <IconButton>
                    <FontAwesomeIcon size="xs" icon={faQuestionCircle} />
                  </IconButton>
                </Tooltip>
              )}
            </Box>
            {content}
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

  return (
    <Layout breadcrumbs={breadcrumbs} user={user}>
      <Box width={sideSheetMode ? SidesheetContentWidth : '100%'}>
        <Box width="100%">
          {!sideSheetMode && (
            <Box width="100%">
              <DetailsPageHeader name={operator ? operator.name : 'Operator'} />
              {operator?.description && (
                <Typography variant="body1">{operator?.description}</Typography>
              )}
            </Box>
          )}
          <Box display="flex" width="100%" pt={sideSheetMode ? '16px' : '40px'}>
            <Box width="100%" mr="32px">
              <ArtifactSummaryList
                title="Inputs"
                workflowId={workflowId}
                dagResultId={workflowDagResultId}
                artifactResults={inputs}
                appearance="link"
              />
            </Box>

            <Box width="100%">
              <ArtifactSummaryList
                title="Outputs"
                workflowId={workflowId}
                dagResultId={workflowDagResultId}
                artifactResults={outputs}
                appearance="link"
              />
            </Box>
          </Box>

          <Divider sx={{ my: '32px' }} />

          <Box>
            <Typography variant="h6" fontWeight="normal">
              Logs
            </Typography>
            {logs !== {} && <LogViewer logs={logs} err={operatorError} />}
          </Box>

          <Divider sx={{ my: '32px' }} />
          {operatorPreview()}
        </Box>
      </Box>
    </Layout>
  );
};

export default OperatorDetailsPage;
