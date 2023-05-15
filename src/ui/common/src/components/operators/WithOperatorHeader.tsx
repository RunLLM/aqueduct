import {
  faCircleExclamation,
  faTriangleExclamation,
} from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Box, Divider, Typography } from '@mui/material';
import React from 'react';
import { useSelector } from 'react-redux';

import { RootState } from '../../stores/store';
import { theme } from '../../styles/theme/theme';
import { OperatorType } from '../../utils/operators';
import { CheckLevel } from '../../utils/operators';
import DetailsPageHeader from '../pages/components/DetailsPageHeader';
import SaveDetails from '../pages/components/SaveDetails';
import ResourceItem from '../pages/workflows/components/ResourceItem';
import ArtifactSummaryList from '../workflows/artifact/summaryList';
import { NodeResultsMap, NodesMap, OperatorResponse, OperatorResultResponse } from '../../handlers/responses/node';

type Props = {
  workflowId: string;
  dagId: string;
  dagResultId?: string;
  nodes: NodesMap;
  operator: OperatorResponse;
  operatorResult?: OperatorResultResponse;
  nodeResults?: NodeResultsMap;
  sideSheetMode?: boolean;
  children?: React.ReactElement | React.ReactElement[];
};

const WithOperatorHeader: React.FC<Props> = ({
  workflowId,
  dagId,
  dagResultId,
  nodes,
  operator,
  operatorResult,
  nodeResults,
  sideSheetMode,
  children,
}) => {
  const integrationsState = useSelector(
    (state: RootState) => state.integrationsReducer
  );

  const operatorStatus = operatorResult?.exec_state?.status;
  const mapArtifacts = (artfIds: string[]) =>
    artfIds
      .map((artifactId) => (nodes ?? {})[artifactId])
      .filter((artf) => !!artf);
  const inputs = mapArtifacts(operator.inputs);
  const outputs = mapArtifacts(operator.outputs);

  let checkLevelDisplay = null;
  if (operator?.spec?.check?.level) {
    const checkLevel = operator.spec.check.level;
    checkLevelDisplay = (
      <Box sx={{ display: 'flex', alignItems: 'center' }} mb={2}>
        <Typography variant="body2" sx={{ color: 'gray.800' }}>
          Check Level
        </Typography>
        <Typography variant="body1" sx={{ mx: 1 }}>
          {checkLevel.charAt(0).toUpperCase() + checkLevel.slice(1)}
        </Typography>
        <FontAwesomeIcon
          icon={
            checkLevel === CheckLevel.Error
              ? faCircleExclamation
              : faTriangleExclamation
          }
          color={
            checkLevel === CheckLevel.Error
              ? theme.palette.red[600]
              : theme.palette.orange[600]
          }
        />
      </Box>
    );
  }

  const service =
    operator?.spec?.load?.service || operator?.spec?.extract?.service;
  const integrationId =
    operator?.spec?.load?.integration_id ||
    operator?.spec?.extract?.integration_id;
  const integrationName = integrationId
    ? integrationsState?.integrations[integrationId]?.name
    : undefined;

  return (
    <Box width="100%">
      {!sideSheetMode && (
        <Box width="100%">
          <DetailsPageHeader
            name={operator ? operator.name : 'Operator'}
            status={operatorStatus}
          />
          {operator?.description && (
            <Typography variant="body1">{operator?.description}</Typography>
          )}
        </Box>
      )}

      <Box width="100%" paddingTop={sideSheetMode ? '16px' : '24px'}>
        {checkLevelDisplay}
      </Box>

      {integrationName && (
        <ResourceItem
          resource={service}
          resourceCustomName={
            integrationsState?.integrations[integrationId]?.name
          }
        />
      )}

      <Box display="flex" width="100%">
        {operator?.spec?.load?.parameters && (
          <Box width="100%" paddingTop={sideSheetMode ? '16px' : '24px'}>
            <SaveDetails parameters={operator?.spec?.load?.parameters} />
          </Box>
        )}

        {operator?.spec?.load && <Box width="96px" />}

        <Box
          display="flex"
          width="100%"
          paddingTop={sideSheetMode ? '16px' : '24px'}
        >
          {inputs.length > 0 && (
            <Box width="100%" mr="32px">
              <ArtifactSummaryList
                title="Inputs"
                workflowId={workflowId}
                dagId={dagId}
                dagResultId={dagResultId}
                artifactResults={inputs}
                collapsePrimitives={operator.spec?.type !== OperatorType.Check}
                appearance={
                  operator.spec?.type === OperatorType.Metric ? 'value' : 'link'
                }
              />
            </Box>
          )}

          {outputs.length > 0 && (
            <Box width="100%">
              <ArtifactSummaryList
                title="Outputs"
                workflowId={workflowId}
                dagId={dagId}
                dagResultId={dagResultId}
                nodes={nodes}
                artifactResults={outputs}
                appearance={
                  operator.spec?.type === OperatorType.Metric ? 'value' : 'link'
                }
              />
            </Box>
          )}
        </Box>
      </Box>

      <Divider sx={{ my: '32px' }} />
      {children}
    </Box>
  );
};

export default WithOperatorHeader;
