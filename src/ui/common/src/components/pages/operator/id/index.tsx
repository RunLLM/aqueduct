import { CircularProgress, Divider } from '@mui/material';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import { BlobReader, TextWriter, ZipReader } from '@zip.js/zip.js';
import React, { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useLocation, useNavigate, useParams } from 'react-router-dom';

import DefaultLayout from '../../../../components/layouts/default';
import { BreadcrumbLink } from '../../../../components/layouts/NavBar';
import LogViewer from '../../../../components/LogViewer';
import MultiFileViewer from '../../../../components/MultiFileViewer';
import { handleGetWorkflowDagResult } from '../../../../handlers/getWorkflowDagResult';
import { AppDispatch, RootState } from '../../../../stores/store';
import UserProfile from '../../../../utils/auth';
import { getPathPrefix } from '../../../../utils/getPathPrefix';
import { exportFunction } from '../../../../utils/operators';
import { LoadingStatusEnum } from '../../../../utils/shared';
import { isInitial, isLoading } from '../../../../utils/shared';
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

  const [defaultFileExtension, setDefaultFileExtension] = useState<string>(".sql");

  const workflowDagResultWithLoadingStatus = useSelector(
    (state: RootState) =>
      state.workflowDagResultsReducer.results[workflowDagResultId]
  );

  const workflow = useSelector((state: RootState) => state.workflowReducer);

  const operator = (workflowDagResultWithLoadingStatus?.result?.operators ??
    {})[operatorId];

  const pathPrefix = getPathPrefix();
  const workflowLink = `${pathPrefix}/workflow/${workflowId}?workflowDagResultId=${workflowDagResultId}`;
  const breadcrumbs = [
    BreadcrumbLink.HOME,
    BreadcrumbLink.WORKFLOWS,
    new BreadcrumbLink(workflowLink, workflow?.selectedDag?.metadata?.name || ''),
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
    console.log('operator changed');
    async function getFilesBlob() {
      console.log('getFilesBlob()');
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

    // TODO: only call this when we are inside a function operator. not an extract or load operator.

    if (operator && operator?.spec?.type !== "extract" && operator?.spec?.type !== "load") {
      console.log('getting files blob');
      getFilesBlob();
    } else if (operator) {
      console.log('getting sql statement from operator');
      let fileContent = '';
      if (operator?.spec?.type === 'extract') {
        fileContent = operator.spec.extract.parameters.query;
      } else if (operator?.spec?.type === 'load') { // TODO

        // TODO: Figure out how to support load operators here.
        // Not sure what to do since load operator object doesn't have a query as part of it.
        /*
          export type LoadParameters = RelationalDBLoadParams | GoogleSheetsLoadParams;

          export type RelationalDBLoadParams = {
            table: string;
            update_mode: string;
          };
        */
        //fileContent = operator.spec.load.parameters.query;
      }
      setFiles((prevState) => {
        const files = {
          // Note: not sure why this is here, but file viewer assumes the first entry to be blank like this
          [""]: {
            path: '',
            language: 'SQL', // or plaintext
            content: ''
          },
          [operator.name]: {
            //[""]: {},
            [`${operator.name}.sql`]: {
              path: `${operator.name}/${operator.name}.sql`,
              language: "sql",
              content: fileContent //operator.spec.parameters.query
            }
          }
        }

        return files;
      });
    }
  }, [operator]);

  useEffect(() => {
    // check operator type to determine what to set the filetype to
    console.log('files changed operator: ', operator);
    if (operator && operator?.spec?.type === 'extract' || operator?.spec?.type === 'load') {
      setDefaultFileExtension('.sql');
    } else if (operator && operator?.spec?.type === 'function') {
      // function operators should default to showing python files.
      setDefaultFileExtension('.py');
    }
  }, [files]);

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

  return (
    <Layout breadcrumbs={breadcrumbs} user={user}>
      <Box width={'800px'}>
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
                initiallyExpanded={true}
              />
            </Box>

            <Box width="100%">
              <ArtifactSummaryList
                title="Outputs"
                workflowId={workflowId}
                dagResultId={workflowDagResultId}
                artifactResults={outputs}
                initiallyExpanded={true}
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

          <Box>
            <Typography variant="h6" fontWeight="normal" mb={1}>
              Code Preview
            </Typography>
            <MultiFileViewer
              files={files}
              defaultFile={operator.name || ''}
              defaultFileExtension={defaultFileExtension}
            />
          </Box>
        </Box>
      </Box>
    </Layout>
  );
};

export default OperatorDetailsPage;
