import {
  faCircleExclamation,
  faTriangleExclamation,
} from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Box, Divider, Typography } from '@mui/material';
import React from 'react';

import { DagResultResponse } from '../../handlers/responses/dag';
import { OperatorResultResponse } from '../../handlers/responses/operator';
import { WorkflowDagResultWithLoadingStatus } from '../../reducers/workflowDagResults';
import { WorkflowDagWithLoadingStatus } from '../../reducers/workflowDags';
import { theme } from '../../styles/theme/theme';
import { OperatorType } from '../../utils/operators';
import { CheckLevel } from '../../utils/operators';
import DetailsPageHeader from '../pages/components/DetailsPageHeader';
import SaveDetails from '../pages/components/SaveDetails';
import ResourceItem from '../pages/workflows/components/ResourceItem';
import ArtifactSummaryList from '../workflows/artifact/summaryList';

type Props = {
  workflowId: string;
  dagId: string;
  dagResultId: string;
  dagWithLoadingStatus?: WorkflowDagWithLoadingStatus;
  dagResultWithLoadingStatus?: WorkflowDagResultWithLoadingStatus;
  operator?: OperatorResultResponse;
  sideSheetMode?: boolean;
  children?: React.ReactElement | React.ReactElement[];
};

const WithOperatorHeader: React.FC<Props> = ({
  workflowId,
  dagId,
  dagResultId,
  dagWithLoadingStatus,
  dagResultWithLoadingStatus,
  operator,
  sideSheetMode,
  children,
}) => {
  if (!operator) {
    return null;
  }

  if (!dagWithLoadingStatus && !dagResultWithLoadingStatus) {
    return null;
  }

  const dagResult =
    dagResultWithLoadingStatus?.result ??
    (dagWithLoadingStatus?.result as DagResultResponse);

  const operatorStatus = operator?.result?.exec_state?.status;
  const mapArtifacts = (artfIds: string[]) =>
    artfIds
      .map((artifactId) => (dagResult?.artifacts ?? {})[artifactId])
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

      <ResourceItem resource={operator?.spec?.load?.service || operator?.spec?.extract?.service} />
      
      <Box display="flex" width="100%">
        <Box width="100%" paddingTop={sideSheetMode ? '16px' : '24px'}>
          <SaveDetails parameters={operator?.spec?.load?.parameters} />
        </Box>

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
