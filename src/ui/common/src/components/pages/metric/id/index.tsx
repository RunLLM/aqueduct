import { faCircleCheck } from '@fortawesome/free-solid-svg-icons';
import { faChevronRight } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Link, List, ListItem } from '@mui/material';
import Accordion from '@mui/material/Accordion';
import AccordionDetails from '@mui/material/AccordionDetails';
import AccordionSummary from '@mui/material/AccordionSummary';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React, { useEffect, useState } from 'react';
import Plot from 'react-plotly.js';
import { useDispatch, useSelector } from 'react-redux';
import { Link as RouterLink, useParams } from 'react-router-dom';

import { boolArtifactNodeIcon } from '../../../../components/workflows/nodes/BoolArtifactNode';
import { checkOperatorNodeIcon } from '../../../../components/workflows/nodes/CheckOperatorNode';
import { databaseNodeIcon } from '../../../../components/workflows/nodes/DatabaseNode';
import { functionOperatorNodeIcon } from '../../../../components/workflows/nodes/FunctionOperatorNode';
import { genericArtifactNodeIcon } from '../../../../components/workflows/nodes/GenericArtifactNode';
import { imageArtifactNodeIcon } from '../../../../components/workflows/nodes/ImageArtifactNode';
import { jsonArtifactNodeIcon } from '../../../../components/workflows/nodes/JsonArtifactNode';
import { metricOperatorNodeIcon } from '../../../../components/workflows/nodes/MetricOperatorNode';
import { numericArtifactNodeIcon } from '../../../../components/workflows/nodes/NumericArtifactNode';
import { stringArtifactNodeIcon } from '../../../../components/workflows/nodes/StringArtifactNode';
import { tableArtifactNodeIcon } from '../../../../components/workflows/nodes/TableArtifactNode';
import { NodeType } from '../../../../reducers/nodeSelection';
import { OperatorResult } from '../../../../reducers/workflow';
import { AppDispatch, RootState } from '../../../../stores/store';
import UserProfile from '../../../../utils/auth';
import { getPathPrefix } from '../../../../utils/getPathPrefix';
import DefaultLayout from '../../../layouts/default';
import { LayoutProps } from '../../types';

type MetricDetailsHeaderProps = {
  artifactName: string;
  createdAt?: string;
};

const MetricDetailsHeader: React.FC<MetricDetailsHeaderProps> = ({
  artifactName,
  createdAt,
}) => {
  return (
    <Box width="100%" display="flex" alignItems="center">
      <FontAwesomeIcon
        height="24px"
        width="24px"
        style={{ marginRight: '8px' }}
        icon={faCircleCheck}
        color={'green'}
      />
      <Typography variant="h4" component="div">
        {artifactName}
      </Typography>
      {createdAt && (
        <Typography marginTop="4px" variant="caption" component="div">
          Created: {createdAt}
        </Typography>
      )}
    </Box>
  );
};

type MetricDetailsPageProps = {
  user: UserProfile;
  Layout?: React.FC<LayoutProps>;
};

const MetricDetailsPage: React.FC<MetricDetailsPageProps> = ({
  user,
  Layout = DefaultLayout,
}) => {
  const dispatch: AppDispatch = useDispatch();
  const { workflowDagResultId, metricOperatorId } = useParams();
  const [inputsExpanded, setInputsExpanded] = useState<boolean>(true);
  const [outputsExpanded, setOutputsExpanded] = useState<boolean>(true);

  const metricResult: OperatorResult | null = useSelector(
    (state: RootState) => {
      // First, check if there are any keys in the operatorResult's object.
      const operatorResults = state.workflowReducer.operatorResults;
      if (Object.keys(operatorResults).length < 1) {
        return null;
      }

      return operatorResults[metricOperatorId];
    }
  );

  useEffect(() => {
    // TODO: Update this to contain the name of the operator
    document.title = 'Metric Details | Aqueduct';

    // if (!metricResult) {
    //     dispatch(
    //         handleGetOperatorResults({
    //             apiKey: user.apiKey,
    //             workflowDagResultId,
    //             operatorId: metricOperatorId,
    //         })
    //     );
    // }
  }, []);

  // TODO: Bring this back after done getting the metricResults.
  // if (!metricResult || !metricResult.result) {
  //     return (
  //         <Layout user={user}>
  //             <CircularProgress />
  //         </Layout>
  //     );
  // }

  //const parsedData = JSON.parse(metricResult.);

  // Set up different metric input types for rendering in the inputs list.
  // TODO: transform/handle response from API and render these appropriately.
  const metricInputs = [
    {
      name: 'bool_artifact',
      nodeType: NodeType.BoolArtifact,
      nodeIcon: boolArtifactNodeIcon,
    },
    {
      name: 'checkOperator',
      nodeType: NodeType.CheckOp,
      nodeIcon: checkOperatorNodeIcon,
    },
    {
      name: 'database_node',
      nodeType: NodeType.LoadOp,
      nodeIcon: databaseNodeIcon,
    },
    {
      name: 'function_operator_node',
      nodeType: NodeType.FunctionOp,
      nodeIcon: functionOperatorNodeIcon,
    },
    {
      name: 'generic_artifact_node',
      nodeType: NodeType.GenericArtifact,
      nodeIcon: genericArtifactNodeIcon,
    },
    {
      name: 'image_artifact_node',
      nodeType: NodeType.ImageArtifact,
      nodeIcon: imageArtifactNodeIcon,
    },
    {
      name: 'jsonArtifactNode',
      nodeType: NodeType.JsonArtifact,
      nodeIcon: jsonArtifactNodeIcon,
    },
    {
      name: 'metric_operator_node',
      nodeType: NodeType.MetricOp,
      nodeIcon: metricOperatorNodeIcon,
    },
    {
      name: 'numeric_artifact_node',
      nodeType: NodeType.NumericArtifact,
      nodeIcon: numericArtifactNodeIcon,
    },
    {
      name: 'string_artifact_node',
      nodeType: NodeType.StringArtifact,
      nodeIcon: stringArtifactNodeIcon,
    },
    {
      name: 'table_artifact_node_icon',
      nodeType: NodeType.TableArtifact,
      nodeIcon: tableArtifactNodeIcon,
    },
  ];

  const listStyle = {
    width: '100%',
    maxWidth: 360,
    bgcolor: 'background.paper',
  };

  return (
    <Layout user={user}>
      <Box width={'800px'}>
        <Box width="100%">
          <Box width="100%" display="flex">
            <MetricDetailsHeader artifactName="metric_result_placeholder" />
          </Box>

          <Box display="flex" width="100%" paddingTop="40px">
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
                  <List sx={listStyle}>
                    {metricInputs.map((metricInput, index) => {
                      return (
                        <ListItem divider key={`metric-input-${index}`}>
                          <Box display="flex">
                            <Box
                              sx={{
                                width: '16px',
                                height: '16px',
                                color: 'rgba(0,0,0,0.54)',
                              }}
                            >
                              <FontAwesomeIcon icon={metricInput.nodeIcon} />
                            </Box>
                            <Link
                              to={`${getPathPrefix()}/workflows`}
                              component={RouterLink as any}
                              sx={{ marginLeft: '16px' }}
                              underline="none"
                            >
                              {metricInput.name}
                            </Link>
                          </Box>
                        </ListItem>
                      );
                    })}
                  </List>
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
                  <Typography variant="h6">125.75</Typography>
                </AccordionDetails>
              </Accordion>
            </Box>
          </Box>

          <Box width="100%" marginTop="12px">
            <Typography variant="h5" component="div" marginBottom="8px">
              Historical Outputs:
            </Typography>
            <Box display="flex" justifyContent="center">
              <Plot
                data={[
                  {
                    x: [1, 2, 3],
                    y: [2, 6, 3],
                    type: 'scatter',
                    mode: 'lines+markers',
                    marker: { color: 'red' },
                  },
                ]}
                layout={{ width: '100%', height: '100%' }}
              />
            </Box>
          </Box>
        </Box>
      </Box>
    </Layout>
  );
};

export default MetricDetailsPage;
