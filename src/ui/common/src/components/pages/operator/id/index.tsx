import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Divider, Link } from '@mui/material';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import { BlobReader, TextWriter, ZipReader } from '@zip.js/zip.js';
import React, { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Link as RouterLink, useNavigate, useParams } from 'react-router-dom';

import DefaultLayout from '../../../../components/layouts/default';
import LogViewer from '../../../../components/LogViewer';
import MultiFileViewer from '../../../../components/MultiFileViewer';
import { artifactTypeToIconMapping } from '../../../../components/workflows/nodes/nodeTypes';
import {
  handleGetArtifactResults,
  handleGetOperatorResults,
  handleGetWorkflow,
  selectResultIdx,
} from '../../../../reducers/workflow';
import { AppDispatch, RootState } from '../../../../stores/store';
import { theme } from '../../../../styles/theme/theme';
import UserProfile from '../../../../utils/auth';
import { getPathPrefix } from '../../../../utils/getPathPrefix';
import { exportFunction } from '../../../../utils/operators';
import { LoadingStatusEnum } from '../../../../utils/shared';
import DetailsPageHeader from '../../components/DetailsPageHeader';
import { LayoutProps } from '../../types';

type OperatorDetailsPageProps = {
  user: UserProfile;
  Layout?: React.FC<LayoutProps>;
  maxRenderSize?: number;
};

// Checked with file size=313285391 and handles that smoothly once loaded. However, takes a while to load.
const OperatorDetailsPage: React.FC<OperatorDetailsPageProps> = ({
  user,
  Layout = DefaultLayout,
  maxRenderSize = 100000000,
}) => {
  const dispatch: AppDispatch = useDispatch();
  const navigate = useNavigate();
  const [files, setFiles] = useState({
    '': {
      path: '',
      language: 'plaintext',
      content: '',
    },
  });

  const params = useParams();
  const workflow = useSelector((state: RootState) => state.workflowReducer);
  const operator = workflow.operatorResults[params.operatorId];
  const inputs =
    workflow.selectedDag?.operators[params.operatorId]?.inputs ?? [];
  const outputs =
    workflow.selectedDag?.operators[params.operatorId]?.outputs ?? [];

  [...inputs, ...outputs].map((artifactId) => {
    if (!workflow.artifactResults[artifactId])
      dispatch(
        handleGetArtifactResults({
          apiKey: user.apiKey,
          workflowDagResultId: params.workflowDagResultId,
          artifactId: artifactId,
        })
      );
  });

  useEffect(() => {
    document.title = 'Operator Details | Aqueduct';
    if (
      workflow.selectedDag === undefined ||
      (workflow.selectedDag && !(params.workflowId in workflow.selectedDag))
    ) {
      dispatch(
        handleGetWorkflow({
          apiKey: user.apiKey,
          workflowId: params.workflowId,
        })
      );
    }
  }, []);

  useEffect(() => {
    if (
      workflow.loadingStatus.loading === LoadingStatusEnum.Succeeded &&
      !(params.operatorId in workflow.operatorResults)
    ) {
      let idx = 0;
      workflow.dagResults.forEach((value, index) => {
        if (value.id === params.workflowDagResultId) {
          idx = index;
        }
      });
      dispatch(selectResultIdx(idx));
      // May encounter a race condition where selectResultIdx sets operatorResults to {}
      // after we populate it because currently cannot check when selectResultIdx is done.
      // Will fix after ui_redesign first pass is done.
      dispatch(
        handleGetOperatorResults({
          apiKey: user.apiKey,
          workflowDagResultId: params.workflowDagResultId,
          operatorId: params.operatorId,
        })
      );
    }
  }, [workflow.loadingStatus.loading]);

  if (operator?.result?.name) {
    document.title = `${operator.result.name} | Aqueduct`;
  }

  // return null if we don't have the workflow loaded.
  // This workflow doesn't exist.
  if (workflow.loadingStatus.loading === LoadingStatusEnum.Failed) {
    navigate('/404');
    return null;
  }

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
      const blob = await exportFunction(user, params.operatorId);
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
    getFilesBlob();
  }, []);

  const mapArtfIds = (artfIds: string[]) => {
    return artfIds.map((artfId, index) => {
      const artifactResult = workflow.artifactResults[artfId];
      if (!artifactResult || !artifactResult.result) {
        return null;
      }

      return (
        <Link
          key={artfId}
          to={`${getPathPrefix()}/workflow/${params.workflowId}/result/${
            params.workflowDagResultId
          }/artifact/${artfId}`}
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
                index === artfIds.length - 1
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
                icon={
                  artifactTypeToIconMapping[artifactResult.result.artifact_type]
                }
              />
            </Box>
            <Typography ml="16px">{artifactResult.result.name}</Typography>
          </Box>
        </Link>
      );
    });
  };

  const inputItems =
    inputs.length > 0 ? (
      mapArtfIds(inputs)
    ) : (
      <Box display="flex" p={1} alignItems="center">
        <Typography height="16px" ml="16px" color="gray.700">
          This operator has no input.
        </Typography>
      </Box>
    );

  const outputItems =
    outputs.length > 0 ? (
      mapArtfIds(outputs)
    ) : (
      <Box display="flex" p={1} alignItems="center">
        <Typography height="16px" ml="16px" color="gray.700">
          This operator has no output.
        </Typography>
      </Box>
    );

  const border = {
    border: '2px',
    borderStyle: 'solid',
    borderRadius: '8px',
    borderColor: 'gray.400',
    margin: '16px',
    padding: '16px',
  };

  return (
    <Layout user={user}>
      <Box width={'800px'}>
        <Box width="100%">
          <Box width="100%">
            <DetailsPageHeader name={operator?.result?.name} />
            {operator?.result?.description && (
              <Typography variant="body1">
                {operator?.result?.description}
              </Typography>
            )}
          </Box>

          <Box display="flex" width="100%" marginTop="64px">
            <Box width="100%" mr="32px">
              <Typography variant="h6" mb="8px" fontWeight="normal">
                Inputs
              </Typography>
              {inputItems}
            </Box>
            <Box width="100%">
              <Typography variant="h6" mb="8px" fontWeight="normal">
                Outputs
              </Typography>
              {outputItems}
            </Box>
          </Box>

          <Divider sx={{ marginY: '32px' }} />

          <Box>
            <Typography variant="h4">Logs</Typography>
            {logs !== {} && (
              <Box sx={border}>
                <LogViewer logs={logs} err={operatorError} />
              </Box>
            )}
          </Box>

          <Box>
            <Typography variant="h4">Code Preview</Typography>
            <Box sx={border}>
              <MultiFileViewer files={files} />
            </Box>
          </Box>
        </Box>
      </Box>
    </Layout>
  );
};

export default OperatorDetailsPage;
