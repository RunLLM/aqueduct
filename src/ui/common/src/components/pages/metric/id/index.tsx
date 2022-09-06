import { faCircleCheck } from '@fortawesome/free-solid-svg-icons';
import { faChevronRight } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { CircularProgress, Link, List, ListItem } from '@mui/material';
import Accordion from '@mui/material/Accordion';
import AccordionDetails from '@mui/material/AccordionDetails';
import AccordionSummary from '@mui/material/AccordionSummary';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React, { useEffect, useState } from 'react';
import Plot from 'react-plotly.js';
import { useDispatch, useSelector } from 'react-redux';
import { Link as RouterLink, useNavigate, useParams } from 'react-router-dom';

import StickyHeaderTable from '../../../../components/tables/StickyHeaderTable';
import { boolArtifactNodeIcon } from '../../../../components/workflows/nodes/BoolArtifactNode';
import { checkOperatorNodeIcon } from '../../../../components/workflows/nodes/CheckOperatorNode';
import { databaseNodeIcon } from '../../../../components/workflows/nodes/DatabaseNode';
import { dictArtifactNodeIcon } from '../../../../components/workflows/nodes/DictArtifactNode';
import { functionOperatorNodeIcon } from '../../../../components/workflows/nodes/FunctionOperatorNode';
import { genericArtifactNodeIcon } from '../../../../components/workflows/nodes/GenericArtifactNode';
import { imageArtifactNodeIcon } from '../../../../components/workflows/nodes/ImageArtifactNode';
import { jsonArtifactNodeIcon } from '../../../../components/workflows/nodes/JsonArtifactNode';
import { metricOperatorNodeIcon } from '../../../../components/workflows/nodes/MetricOperatorNode';
import { numericArtifactNodeIcon } from '../../../../components/workflows/nodes/NumericArtifactNode';
import { stringArtifactNodeIcon } from '../../../../components/workflows/nodes/StringArtifactNode';
import { tableArtifactNodeIcon } from '../../../../components/workflows/nodes/TableArtifactNode';
import { NodeType } from '../../../../reducers/nodeSelection';
import {
  handleGetArtifactResults,
  handleGetOperatorResults,
  handleGetWorkflow,
} from '../../../../reducers/workflow';
import { AppDispatch, RootState } from '../../../../stores/store';
import { ArtifactType } from '../../../../utils/artifacts';
import UserProfile from '../../../../utils/auth';
import { Data } from '../../../../utils/data';
import { getPathPrefix } from '../../../../utils/getPathPrefix';
import { LoadingStatusEnum } from '../../../../utils/shared';
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
  const navigate = useNavigate();
  const { workflowId, workflowDagResultId, metricOperatorId } = useParams();

  const [inputsExpanded, setInputsExpanded] = useState<boolean>(true);
  const [outputsExpanded, setOutputsExpanded] = useState<boolean>(true);

  const workflow = useSelector((state: RootState) => state.workflowReducer);
  const workflowDag = workflow.dagResults.find((currentDag) => {
    return currentDag.id === workflowDagResultId;
  });

  const workflowDagId = workflowDag?.workflow_dag_id;
  const operator = (workflow.dags[workflowDagId]?.operators ?? {})[
    metricOperatorId
  ];

  // // Get the operatorSpec so that we can show more metadata about the operator.
  // const operatorSpec = operator?.spec;
  // console.log('operatorSpec: ', operatorSpec);

  // const logs =
  //   workflow.operatorResults[metricOperatorId]?.result?.exec_state?.user_logs ??
  //   {};
  // console.log('logs: ', logs);
  // const operatorError =
  //   workflow?.operatorResults[metricOperatorId]?.result?.exec_state?.error;
  // console.log('operatorError', operatorError);

  // // Get the execution state of the operator:
  // // TODO: Show execution state in a badge component next to title of operator.
  // const execState: ExecState =
  //   workflow?.operatorResults[metricOperatorId]?.result?.exec_state;
  // console.log('execState: ', execState);

  const metricResult = workflow.operatorResults[metricOperatorId];

  useEffect(() => {
    // TODO: Update this to contain the name of the operator
    document.title = 'Metric Details | Aqueduct';
    dispatch(handleGetWorkflow({ apiKey: user.apiKey, workflowId }));
  }, []);

  useEffect(() => {
    if (workflow.dagResults.length > 0) {
      dispatch(
        handleGetOperatorResults({
          apiKey: user.apiKey,
          workflowDagResultId,
          operatorId: metricOperatorId,
        })
      );
    }
  }, [workflow.dagResults]);

  // Set up different metric input types for rendering in the inputs list.
  // TODO: transform/handle response from API and render these appropriately.
  const mockMetricInputs = [
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

  const listStyle = {
    width: '100%',
    maxWidth: 360,
    bgcolor: 'background.paper',
  };

  // return null if we don't have the workflow loaded.
  // This workflow doesn't exist.
  if (workflow.loadingStatus.loading === LoadingStatusEnum.Failed) {
    navigate('/404');
    return null;
  }

  if (operator) {
    const inputs = operator.inputs;
    const outputs = operator.outputs;

    const operatorArtifacts = [...inputs, ...outputs];
    // fetch output artifacts.
    operatorArtifacts.map((artifactId) => {
      if (!workflow.artifactResults[artifactId])
        dispatch(
          handleGetArtifactResults({
            apiKey: user.apiKey,
            workflowDagResultId,
            artifactId,
          })
        );
    });
  }

  if (!metricResult || !metricResult.result) {
    return (
      <Layout user={user}>
        <CircularProgress />
      </Layout>
    );
  }

  // Function to get the numerical value of the metric output
  const getOperatorOutput = () => {
    if (!operator || !operator.outputs) {
      return <CircularProgress />;
    }

    return operator.outputs.map((artifactId) => {
      const artifactResult = workflow.artifactResults[artifactId];
      if (!artifactResult) {
        return null;
      }

      // TODO: Check the serialization_type of the artifacts and show a link
      // to table vew artifacts when present.
      // TODO: Do the same thing that we're doing over in inputs :)

      if (artifactResult.result) {
        if (artifactResult.result.artifact_type === 'table') {
          // Link to appropriate artifact details page
          // Show tableIcon here as part of the link.
          return (
            <Box key={artifactId}>
              <Typography variant="body1">TEST</Typography>
            </Box>
          );
        } else {
          // Render inline if possible
          return (
            <Box>
              <Typography variant="body1">
                {artifactResult.result.data}
              </Typography>
            </Box>
          );
        }
      }

      return null;
    });
  };

  // TODO: refactor getOperatorInput and getOperatorOutput to just one function.
  // was playing around with design before i figured they can be the same.
  const getOperatorInput = () => {
    if (!operator || !operator.inputs) {
      return <CircularProgress />;
    }

    return operator.inputs.map((artifactId, index) => {
      const artifactResult = workflow.artifactResults[artifactId];
      if (!artifactResult) {
        return null;
      }

      if (artifactResult.result) {
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
                <FontAwesomeIcon
                  icon={
                    artifactTypeToIconMapping[
                      artifactResult.result.artifact_type
                    ]
                  }
                />
              </Box>
              <Link
                to={`${getPathPrefix()}/workflow/${workflowId}/result/${workflowDagResultId}/artifact/${artifactId}`}
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

  // Mock out the historical metrics here.
  const mockHistoricalMetrics: Data = {
    schema: {
      fields: [
        { name: 'status', type: 'varchar' },
        { name: 'timestamp', type: 'varchar' },
        { name: 'value', type: 'float' },
      ],
      pandas_version: '0.0.1',
    },
    data: [
      { status: 'Succeeded', timestamp: '03/14/2022 04:00 PST', value: 124.5 },
      { status: 'Succeeded', timestamp: '03/15/2022 04:00 PST', value: 128.5 },
      { status: 'Warning', timestamp: '03/16/2022 04:00 PST', value: 127.5 },
      { status: 'Error', timestamp: '03/17/2922 04:00 PST', value: 100 },
    ],
  };

  const mockHistoricalMetricTimestamps = mockHistoricalMetrics.data.map(
    (mockHistoricalData) => mockHistoricalData.timestamp
  );
  const mockHistoricalMetricValues = mockHistoricalMetrics.data.map(
    (mockHistoricalData) => mockHistoricalData.value
  );

  return (
    <Layout user={user}>
      <Box width={'800px'}>
        <Box width="100%">
          <Box width="100%">
            <MetricDetailsHeader artifactName={metricResult.result.name} />
            {metricResult.result?.description && (
              <Typography variant="body1">
                {metricResult.result.description}
              </Typography>
            )}
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

          <Box width="100%" marginTop="12px">
            <Typography variant="h5" component="div" marginBottom="8px">
              Historical Outputs:
            </Typography>
            <Box display="flex" justifyContent="center" flexDirection="column">
              <Plot
                data={[
                  {
                    x: mockHistoricalMetricTimestamps,
                    y: mockHistoricalMetricValues,
                    type: 'scatter',
                    mode: 'lines+markers',
                    marker: { color: 'red' },
                  },
                ]}
                layout={{ width: '100%', height: '100%' }}
              />
              <StickyHeaderTable data={mockHistoricalMetrics} />
            </Box>
          </Box>
        </Box>
      </Box>
    </Layout>
  );
};

export default MetricDetailsPage;
