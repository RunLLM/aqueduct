import { faGear, faPlay } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Alert, Snackbar, Typography } from '@mui/material';
import Box from '@mui/material/Box';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import TextField from '@mui/material/TextField';
import React, { useState } from 'react';
import ReactMarkdown from 'react-markdown';
import { useSelector } from 'react-redux';
import { useDispatch } from 'react-redux';
import { useNavigate } from 'react-router-dom';

import { handleLoadIntegrations } from '../../reducers/integrations';
import { handleGetWorkflow, selectResultIdx } from '../../reducers/workflow';
import { RootState } from '../../stores/store';
import { AppDispatch } from '../../stores/store';
import style from '../../styles/markdown.module.css';
import {
  Artifact,
  ArtifactType,
  SerializationType,
} from '../../utils/artifacts';
import UserProfile from '../../utils/auth';
import { getNextUpdateTime } from '../../utils/cron';
import { EngineType } from '../../utils/engine';
import { WorkflowDag, WorkflowUpdateTrigger } from '../../utils/workflows';
import { useAqueductConsts } from '../hooks/useAqueductConsts';
import { Button } from '../primitives/Button.styles';
import { WorkflowStatusBar } from './StatusBar';
import VersionSelector from './version_selector';
import WorkflowSettings from './WorkflowSettings';
import StatusChip from './workflowStatus';

export const WorkflowPageContentId = 'workflow-page-main';

type Props = {
  user: UserProfile;
  workflowDag: WorkflowDag;
  workflowId: string;
};

const ContainerWidthBreakpoint = 700;

const WorkflowHeader: React.FC<Props> = ({ user, workflowDag, workflowId }) => {
  const dispatch: AppDispatch = useDispatch();
  const { apiAddress } = useAqueductConsts();
  const navigate = useNavigate();

  const currentNode = useSelector(
    (state: RootState) => state.nodeSelectionReducer.selected
  );

  // NOTE: The 1000 here is just a placeholder. By the time the page snaps into place,
  // it will be overridden.
  const [containerWidth, setContainerWidth] = useState(1000);
  const narrowView = containerWidth < ContainerWidthBreakpoint;

  const getContainerSize = () => {
    const container = document.getElementById(WorkflowPageContentId);

    if (!container) {
      // The page hasn't fully rendered yet.
      setContainerWidth(1000); // Just a default value.
    } else {
      setContainerWidth(container.clientWidth);
    }
  };

  // TODO (ENG-2302): useLayoutEffect here. May want to figure out some way to debounce as it gets called quite quickly when resizing.
  window.addEventListener('resize', getContainerSize);

  const [showRunWorkflowDialog, setShowRunWorkflowDialog] = useState(false);
  const workflow = useSelector((state: RootState) => state.workflowReducer);

  const successMessage =
    'Successfully triggered a manual update for this workflow!';
  const [errorMessage, setErrorMessage] = useState(
    'Unable to update this workflow.'
  );

  const [showErrorToast, setShowErrorToast] = useState(false);
  const [showSuccessToast, setShowSuccessToast] = useState(false);
  const [showSettings, setShowSettings] = useState(false);

  const handleSuccessToastClose = async () => {
    setShowSuccessToast(false);

    try {
      await dispatch(handleGetWorkflow({ apiKey: user.apiKey, workflowId }));
      await dispatch(handleLoadIntegrations({ apiKey: user.apiKey }));
      dispatch(selectResultIdx(0));
      navigate(`/workflow/${workflowId}`, { replace: true });
    } catch (error) {
      setErrorMessage(
        `We're having trouble getting the latest workflow. Please try refreshing the page.`
      );
      setShowErrorToast(true);
    }
  };

  const handleErrorToastClose = () => {
    setShowErrorToast(false);
  };

  const name = workflowDag.metadata?.name ?? '';
  const description = workflowDag.metadata?.description;

  let nextUpdateComponent;
  if (
    workflowDag.metadata?.schedule?.trigger ===
      WorkflowUpdateTrigger.Periodic &&
    !workflowDag.metadata?.schedule?.paused
  ) {
    const nextUpdateTime = getNextUpdateTime(
      workflowDag.metadata?.schedule?.cron_schedule
    );
    nextUpdateComponent = (
      <Box sx={{ mt: 1 }}>
        <Typography variant="body2">
          <strong> Workflow Engine: </strong>
          {workflowDag.engine_config.type}
        </Typography>
        <Typography variant="body2">
          <strong> Next Workflow Run: </strong>{' '}
          {nextUpdateTime.toDate().toLocaleString()}
        </Typography>
      </Box>
    );
  }

  const showAirflowUpdateWarning =
    workflowDag.engine_config.type === EngineType.Airflow &&
    !workflowDag.engine_config.airflow_config?.matches_airflow;
  const airflowUpdateWarning = (
    <Box maxWidth="800px">
      <Alert severity="warning">
        Please copy the latest Airflow DAG file to your Airflow server if you
        have not done so already. New Airflow DAG runs will not be synced
        properly with Aqueduct until you have copied the file.
      </Alert>
    </Box>
  );

  const paramNameToDisplayProps = Object.assign(
    {},
    ...Object.values(workflowDag.operators)
      .filter((operator) => {
        return operator.spec.param !== undefined;
      })
      .map((operator) => {
        // Parameter operators should only have a single output.
        if (operator.outputs.length > 1) {
          console.error('Parameter operator should not have multiple outputs.');
        }

        // Some types of parameters cannot be easily customized from a textfield on the UI.
        // These types are not json-able and cannot be easily typed as strings.
        const outputArtifact: Artifact =
          workflowDag.artifacts[operator.outputs[0]];
        const isCustomizable = ![
          ArtifactType.Table,
          ArtifactType.Bytes,
          ArtifactType.Tuple,
          ArtifactType.Image,
          ArtifactType.Picklable,
        ].includes(outputArtifact.type);

        let placeholder: string;
        let helperText: string;
        if (isCustomizable) {
          placeholder = atob(operator.spec.param.val);
          helperText = '';
        } else {
          placeholder = '';
          helperText =
            outputArtifact.type[0].toUpperCase() +
            outputArtifact.type.substr(1) +
            ' type is not yet customizable from the UI.';
        }

        return {
          [operator.name]: {
            placeholder: placeholder,
            isCustomizable: isCustomizable,
            helperText: helperText,
          },
        };
      })
  );

  // This records all the parameters and values that the user wants to overwrite with.
  const [paramNameToValMap, setParamNameToValMap] = useState<{
    [key: string]: string;
  }>({});

  // Returns the map of parameters, from name to spec (which includes the base64-encoded
  // value and serialization_type).
  const serializeParameters = () => {
    const serializedParams = {};
    Object.entries(paramNameToValMap).forEach(([key, strVal]) => {
      // Serialize the user's input string appropriately into base64. The input can either be a
      // 1) number 2) string 3) json.
      try {
        // All jsonable values are serialized as json.
        JSON.parse(strVal);
        serializedParams[key] = {
          val: btoa(strVal),
          serialization_type: SerializationType.Json,
        };
      } catch (err) {
        // Non-jsonable values (such as plain strings) are serialized as strings.
        serializedParams[key] = {
          val: btoa(strVal),
          serialization_type: SerializationType.String,
        };
      }
    });
    return serializedParams;
  };

  const triggerWorkflowRun = () => {
    const parameters = new FormData();
    parameters.append('parameters', JSON.stringify(serializeParameters()));

    setShowRunWorkflowDialog(false);

    fetch(`${apiAddress}/api/workflow/${workflowDag.workflow_id}/refresh`, {
      method: 'POST',
      headers: {
        'api-key': user.apiKey,
      },
      body: parameters,
    })
      .then((res) => {
        res.json().then((body) => {
          if (res.ok) {
            setShowSuccessToast(true);
          } else {
            setErrorMessage(`Unable to run this workflow: ${body.error}`);
            setShowErrorToast(true);
          }
        });
      })
      .catch(() => {
        setShowErrorToast(true);
      });

    // Reset the overriding parameters map on dialog close.
    setParamNameToValMap({});
  };

  const runWorkflowDialog = (
    <Dialog
      open={showRunWorkflowDialog}
      onClose={() => setShowRunWorkflowDialog(false)}
    >
      <DialogTitle>Trigger a Workflow Run?</DialogTitle>
      <DialogContent>
        <Box sx={{ mb: 2 }}>
          This will trigger a run of <code>{name}</code> immediately.
        </Box>

        {Object.keys(paramNameToDisplayProps).length > 0 && (
          <Box>
            <Typography sx={{ mb: 1 }} style={{ fontWeight: 'bold' }}>
              {' '}
              Parameters{' '}
            </Typography>
            <Typography variant="caption">
              For json-serializable types like dictionaries or lists, enter the
              string-serialized representation, without the outer quotes. That
              is to say, the result of `json.dumps(val)`.
            </Typography>
          </Box>
        )}
        {Object.keys(paramNameToDisplayProps).map((paramName) => {
          return (
            <Box key={paramName}>
              <Typography>
                <small>{paramName}</small>
              </Typography>
              <TextField
                fullWidth
                disabled={!paramNameToDisplayProps[paramName].isCustomizable}
                helperText={paramNameToDisplayProps[paramName].helperText}
                placeholder={paramNameToDisplayProps[paramName].placeholder}
                onChange={(e) => {
                  paramNameToValMap[paramName] = e.target.value;
                  setParamNameToValMap(paramNameToValMap);
                }}
                size="small"
              />
            </Box>
          );
        })}
      </DialogContent>
      <DialogActions>
        <Button
          color="secondary"
          onClick={() => setShowRunWorkflowDialog(false)}
        >
          Cancel
        </Button>
        <Button color="primary" onClick={() => triggerWorkflowRun()}>
          Run
        </Button>
      </DialogActions>
    </Dialog>
  );

  return (
    <Box>
      <Box
        sx={{
          display: 'flex',
          alignItems: narrowView ? 'start' : 'center',
          flexDirection: narrowView ? 'column' : 'row',
        }}
      >
        <Box sx={{ flex: 1 }}>
          <StatusChip status={workflow.dagResults[0].status} />

          <Typography
            variant="h4"
            sx={{ my: 1, lineHeight: 1 }}
            fontWeight="normal"
          >
            {name}
          </Typography>
        </Box>

        <Box sx={{ mr: 4 }}>
          <Button
            color="primary"
            sx={{ height: '100%', mr: 2, fontSize: '20px' }}
            onClick={() => setShowRunWorkflowDialog(true)}
            size="large"
          >
            <FontAwesomeIcon icon={faPlay} />
            <Typography sx={{ ml: 1 }}>Run Workflow</Typography>
          </Button>

          <Button
            variant="outlined"
            color="primary"
            onClick={() => setShowSettings(true)}
            sx={{ py: 0 }}
          >
            <Box sx={{ fontSize: '20px' }}>
              <FontAwesomeIcon icon={faGear} />
            </Box>
          </Button>

          <WorkflowSettings
            user={user}
            open={showSettings}
            onClose={() => setShowSettings(false)}
            workflowDag={workflowDag}
          />
        </Box>
      </Box>

      {description && (
        <Typography variant="body1">
          <ReactMarkdown className={style.reactMarkdown}>
            {description}
          </ReactMarkdown>
        </Typography>
      )}

      {nextUpdateComponent}

      <Box
        sx={{
          display: 'flex',
          alignItems: narrowView ? 'start' : 'center',
          my: 1,
          flexDirection: narrowView ? 'column' : 'row',
        }}
      >
        {workflow.dagResults && workflow.dagResults.length > 0 && (
          <VersionSelector />
        )}

        <Box mx={narrowView ? 0 : 2} my={narrowView ? 1 : 0}>
          <WorkflowStatusBar user={user} />
        </Box>
      </Box>

      {runWorkflowDialog}

      {showAirflowUpdateWarning && airflowUpdateWarning}

      <Snackbar
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
        open={showSuccessToast}
        onClose={handleSuccessToastClose}
        key={'workflowheader-success-snackbar'}
        autoHideDuration={4000}
      >
        <Alert
          onClose={handleSuccessToastClose}
          severity="success"
          sx={{ width: '100%' }}
        >
          {successMessage}
        </Alert>
      </Snackbar>
      <Snackbar
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
        open={showErrorToast}
        onClose={handleErrorToastClose}
        key={'workflowheader-error-snackbar'}
        autoHideDuration={4000}
      >
        <Alert
          onClose={handleErrorToastClose}
          severity="error"
          sx={{ width: '100%' }}
        >
          {errorMessage}
        </Alert>
      </Snackbar>
    </Box>
  );
};

export default WorkflowHeader;
