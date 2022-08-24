import { faCircleCheck } from '@fortawesome/free-solid-svg-icons';
import { faChevronRight } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import Accordion from '@mui/material/Accordion';
import AccordionDetails from '@mui/material/AccordionDetails';
import AccordionSummary from '@mui/material/AccordionSummary';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React, { useEffect, useState } from 'react';
import Plot from 'react-plotly.js';
import { useDispatch, useSelector } from 'react-redux';
import { useParams } from 'react-router-dom';

import { OperatorResult } from '../../../../reducers/workflow';
import { AppDispatch, RootState } from '../../../../stores/store';
import UserProfile from '../../../../utils/auth';
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
  const [inputsExpanded, setInputsExpanded] = useState<boolean>(false);
  const [outputsExpanded, setOutputsExpanded] = useState<boolean>(false);

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
                  <Typography>
                    Nulla facilisi. Phasellus sollicitudin nulla et quam mattis
                    feugiat. Aliquam eget maximus est, id dignissim quam.
                  </Typography>
                </AccordionDetails>
              </Accordion>
            </Box>
            <Box width="32px" />
            <Box width="100%">
              <Accordion
                expanded={outputsExpanded}
                onChange={() => {
                  console.log('on accordion change');
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

          {/* <Box display="flex" width="100%" paddingTop="40px">
                        <Box width="100%">
                            <Typography variant="h5" component="div" marginBottom="8px">
                                Metrics
                            </Typography>
                            <KeyValueTable />
                        </Box>
                        <Box width="96px" />
                        <Box width="100%">
                            <Typography variant="h5" component="div" marginBottom="8px">
                                Checks
                            </Typography>
                            <KeyValueTable />
                        </Box>
                    </Box> */}

          <Box width="100%" marginTop="12px">
            <Typography variant="h5" component="div" marginBottom="8px">
              Historical Outputs:
            </Typography>
            <Box>
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
