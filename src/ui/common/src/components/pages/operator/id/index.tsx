import { faChevronRight } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Link, List, ListItem } from '@mui/material';
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

type OperatorDetailsPageProps = {
  user: UserProfile;
  Layout?: React.FC<LayoutProps>;
  maxRenderSize?: number;
};

const listStyle = {
  width: '100%',
  maxWidth: 360,
  bgcolor: 'background.paper',
};

// Checked with file size=313285391 and handles that smoothly once loaded. However, takes a while to load.
const OperatorDetailsPage: React.FC<OperatorDetailsPageProps> = ({
  user,
  Layout = DefaultLayout,
  maxRenderSize = 100000000,
}) => {
  const dispatch: AppDispatch = useDispatch();
  const navigate = useNavigate();
  const [inputsExpanded, setInputsExpanded] = useState<boolean>(true);
  const [outputsExpanded, setOutputsExpanded] = useState<boolean>(true);
  const [files, setFiles] = useState({
    '': {
      path: '',
      language: 'plaintext',
      content: '',
    },
  });

  const artifactTypeToIconMapping = {
    [ArtifactType.String]: stringArtifactNodeIcon,
    [ArtifactType.Bool]: boolArtifactNodeIcon,
    [ArtifactType.Numeric]: numericArtifactNodeIcon,
    [ArtifactType.Dict]: dictArtifactNodeIcon,
    // TODO: figure out if we should use other icon for tuple
    [ArtifactType.Tuple]: dictArtifactNodeIcon,
    [ArtifactType.Table]: tableArtifactNodeIcon,
    [ArtifactType.Json]: jsonArtifactNodeIcon,
    // TODO: figure out what to show for bytes.
    [ArtifactType.Bytes]: dictArtifactNodeIcon,
    [ArtifactType.Image]: imageArtifactNodeIcon,
    // TODO: Figure out what to show for Picklable
    [ArtifactType.Picklable]: dictArtifactNodeIcon,
  };

  const params = useParams();
  const workflow = useSelector((state: RootState) => state.workflowReducer);
  const operator = workflow.operatorResults[params.operatorId];
  const inputs = workflow.selectedDag?.operators[params.operatorId]?.inputs;
  const outputs = workflow.selectedDag?.operators[params.operatorId]?.outputs;
  if (inputs && outputs) {
    const operatorArtifacts = [...inputs, ...outputs];
    // fetch output artifacts.
    operatorArtifacts.map((artifactId) => {
      if (!workflow.artifactResults[artifactId])
        dispatch(
          handleGetArtifactResults({
            apiKey: user.apiKey,
            workflowDagResultId: params.workflowDagResultId,
            artifactId: artifactId,
          })
        );
    });
  }

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

  const getOperatorInput = () => {
    if (!inputs) {
      return null;
    }
    return inputs.map((artifactId, index) => {
      const artifactResult = workflow.artifactResults[artifactId];
      if (!artifactResult) {
        return null;
      }

      if (artifactResult.result) {
        return (
          <ListItem divider key={`fn-input-${index}`}>
            <Box display="flex">
              <Box
                sx={{
                  width: '16px',
                  height: '16px',
                  color: 'rgba(0,0,0,0.54)',
                }}
              >
                <FontAwesomeIcon
                  icon={
                    artifactTypeToIconMapping[
                      artifactResult.result.artifact_type
                    ]
                  }
                />
              </Box>
              <Link
                to={`${getPathPrefix()}/workflow/${params.workflowId}/result/${
                  params.workflowDagResultId
                }/artifact/${artifactId}`}
                component={RouterLink as any}
                sx={{ marginLeft: '16px' }}
                underline="none"
              >
                {artifactResult.result.name}
              </Link>
            </Box>
          </ListItem>
        );
      }

      return null;
    });
  };
  const getOperatorOutput = () => {
    if (!outputs) {
      return null;
    }
    return outputs.map((artifactId, index) => {
      const artifactResult = workflow.artifactResults[artifactId];
      if (!artifactResult) {
        return null;
      }

      if (artifactResult.result) {
        return (
          <ListItem divider key={`fn-output-${index}`}>
            <Box display="flex">
              <Box
                sx={{
                  width: '16px',
                  height: '16px',
                  color: 'rgba(0,0,0,0.54)',
                }}
              >
                <FontAwesomeIcon
                  icon={
                    artifactTypeToIconMapping[
                      artifactResult.result.artifact_type
                    ]
                  }
                />
              </Box>
              <Link
                to={`${getPathPrefix()}/workflow/${params.workflowId}/result/${
                  params.workflowDagResultId
                }/artifact/${artifactId}`}
                component={RouterLink as any}
                sx={{ marginLeft: '16px' }}
                underline="none"
              >
                {artifactResult.result.name}
              </Link>
            </Box>
          </ListItem>
        );
      }

      return null;
    });
  };

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

          <Box
            display="flex"
            width="100%"
            paddingTop="40px"
            paddingBottom="40px"
          >
            <Box width="100%">
              <Accordion
                expanded={inputsExpanded}
                onChange={() => {
                  setInputsExpanded(!inputsExpanded);
                }}
              >
                <AccordionSummary
                  expandIcon={<FontAwesomeIcon icon={faChevronRight} />}
                  aria-controls="input-accordion-content"
                  id="input-accordion-header"
                >
                  <Typography
                    sx={{ width: '33%', flexShrink: 0 }}
                    variant="h5"
                    component="div"
                    marginBottom="8px"
                  >
                    Inputs:
                  </Typography>
                </AccordionSummary>
                <AccordionDetails>
                  <List sx={listStyle}>{getOperatorInput()}</List>
                </AccordionDetails>
              </Accordion>
            </Box>
            <Box width="32px" />
            <Box width="100%">
              <Accordion
                expanded={outputsExpanded}
                onChange={() => {
                  setOutputsExpanded(!outputsExpanded);
                }}
              >
                <AccordionSummary
                  expandIcon={<FontAwesomeIcon icon={faChevronRight} />}
                  aria-controls="panel1bh-content"
                  id="panel1bh-header"
                >
                  <Typography
                    sx={{ width: '33%', flexShrink: 0 }}
                    variant="h5"
                    component="div"
                    marginBottom="8px"
                  >
                    Output:
                  </Typography>
                </AccordionSummary>
                <AccordionDetails>
                  <React.Fragment>{getOperatorOutput()}</React.Fragment>
                </AccordionDetails>
              </Accordion>
            </Box>
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
