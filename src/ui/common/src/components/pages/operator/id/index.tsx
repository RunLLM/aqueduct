import { faChevronRight } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { CircularProgress, Link, List, ListItem } from '@mui/material';
import Accordion from '@mui/material/Accordion';
import AccordionDetails from '@mui/material/AccordionDetails';
import AccordionSummary from '@mui/material/AccordionSummary';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import { BlobReader, TextWriter, ZipReader } from '@zip.js/zip.js';
import React, { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Link as RouterLink, useNavigate, useParams } from 'react-router-dom';

import DefaultLayout from '../../../../components/layouts/default';
import LogViewer from '../../../../components/LogViewer';
import MultiFileViewer from '../../../../components/MultiFileViewer';
import { boolArtifactNodeIcon } from '../../../../components/workflows/nodes/BoolArtifactNode';
import { dictArtifactNodeIcon } from '../../../../components/workflows/nodes/DictArtifactNode';
import { imageArtifactNodeIcon } from '../../../../components/workflows/nodes/ImageArtifactNode';
import { jsonArtifactNodeIcon } from '../../../../components/workflows/nodes/JsonArtifactNode';
import { numericArtifactNodeIcon } from '../../../../components/workflows/nodes/NumericArtifactNode';
import { stringArtifactNodeIcon } from '../../../../components/workflows/nodes/StringArtifactNode';
import { tableArtifactNodeIcon } from '../../../../components/workflows/nodes/TableArtifactNode';
import {
  handleGetArtifactResults,
  handleGetOperatorResults,
  handleGetWorkflow,
  selectResultIdx,
} from '../../../../reducers/workflow';
import { AppDispatch, RootState } from '../../../../stores/store';
import { ArtifactType } from '../../../../utils/artifacts';
import UserProfile from '../../../../utils/auth';
import { getPathPrefix } from '../../../../utils/getPathPrefix';
import { exportFunction } from '../../../../utils/operators';
import { LoadingStatusEnum } from '../../../../utils/shared';
import DetailsPageHeader from '../../components/DetailsPageHeader';
import { LayoutProps } from '../../types';
import ArtifactSummaryList from '../../../workflows/artifact/summaryList';
import { ArtifactResultResponse } from 'src/handlers/responses/artifact';
import { isFailed, isInitial, isLoading } from '../../../../utils/shared';
import { handleGetWorkflowDagResult } from '../../../../handlers/getWorkflowDagResult';
import { handleListArtifactResults } from '../../../../handlers/listArtifactResults';

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
  const { workflowId, workflowDagResultId, operatorId } = useParams();

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

  useEffect(() => {
    document.title = 'Operator Details | Aqueduct';
    
    if ( // Load workflow dag result if it's not cached
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
    if (!!operator) {
      document.title = `${operator.name} | Aqueduct`;
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
    getFilesBlob();
  }, []);

  if (
    !workflowDagResultWithLoadingStatus ||
    isInitial(workflowDagResultWithLoadingStatus.status) ||
    isLoading(workflowDagResultWithLoadingStatus.status)
  ) {
    return (
      <Layout user={user}>
        <CircularProgress />
      </Layout>
    );
  }

  // This workflow doesn't exist.
  if (workflowDagResultWithLoadingStatus.status.loading === LoadingStatusEnum.Failed) {
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
            <DetailsPageHeader name={operator?.name} />
            {operator?.description && (
              <Typography variant="body1">
                {operator?.description}
              </Typography>
            )}
          </Box>

          <Box
            display="flex"
            width="100%"
            paddingTop="40px"
            paddingBottom="40px"
          >
            <ArtifactSummaryList 
              title="Inputs"
              workflowId={workflowId}
              dagResultId={workflowDagResultId}
              artifactResults={inputs}
              initiallyExpanded={true}
            />
            
            <ArtifactSummaryList 
              title="Outputs"
              workflowId={workflowId}
              dagResultId={workflowDagResultId}
              artifactResults={outputs}
              initiallyExpanded={true}
            />
          </Box>

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
