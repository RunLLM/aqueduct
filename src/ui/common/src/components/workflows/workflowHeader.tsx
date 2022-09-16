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
import UserProfile from '../../utils/auth';
import { getNextUpdateTime } from '../../utils/cron';
import { WorkflowDag, WorkflowUpdateTrigger } from '../../utils/workflows';
import { useAqueductConsts } from '../hooks/useAqueductConsts';
import { Button } from '../primitives/Button.styles';
import VersionSelector from './version_selector';
import WorkflowSettings from './WorkflowSettings';
import Status from './workflowStatus';

type Props = {
  user: UserProfile;
  workflowDag: WorkflowDag;
  workflowId: string;
};

const WorkflowHeader: React.FC<Props> = ({ user, workflowDag, workflowId }) => {
  const dispatch: AppDispatch = useDispatch();
  const { apiAddress } = useAqueductConsts();
  const navigate = useNavigate();

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
      setErrorMessage(`We're having trouble getting the latest workflow. Please try refreshing the page.`);
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
          <strong> Next Workflow Run: </strong>{' '}
          {nextUpdateTime.toDate().toLocaleString()}
        </Typography>
      </Box>
    );
  }

  const paramNameToDefault = Object.assign(
    {},
    ...Object.values(workflowDag.operators)
      .filter((operator) => {
        return operator.spec.param !== undefined;
      })
      .map((operator) => {
        return { [operator.name]: operator.spec.param.val };
      })
  );

  // This records all the parameters and values that the user wants to overwrite with.
  const [paramNameToValMap, setParamNameToValMap] = useState<{
    [key: string]: string;
  }>({});

  const triggerWorkflowRun = () => {
    const parameters = new FormData();
    parameters.append('parameters', JSON.stringify(paramNameToValMap));

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

        {Object.keys(paramNameToDefault).length > 0 && (
          <Typography sx={{ mb: 1 }} style={{ fontWeight: 'bold' }}>
            {' '}
            Parameters{' '}
          </Typography>
        )}
        {Object.keys(paramNameToDefault).map((paramName) => {
          return (
            <Box key={paramName}>
              <Typography>
                <small>{paramName}</small>
              </Typography>
              <TextField
                fullWidth
                placeholder={paramNameToDefault[paramName]}
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
      <Box sx={{ display: 'flex', alignItems: 'center', my: 1 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', flex: 1 }}>
          <Typography
            variant="h3"
            sx={{ fontFamily: 'Monospace', mr: 2, lineHeight: 1 }}
          >
            {name}
          </Typography>

          <Status status={workflow.dagResults[0].status} />
        </Box>

        <Box sx={{ ml: 2 }}>
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

      <Box sx={{ display: 'flex', alignItems: 'center', my: 1 }}>
        {workflow.dagResults && workflow.dagResults.length > 0 && (
          <VersionSelector />
        )}

        {/* NOTE: Funnyily enough, `size=large` on a button is what
                    makes it match the size of the `FormControl` when set to
                    small. Go figure. */}
        <Button
          color="primary"
          sx={{ height: '100%' }}
          onClick={() => setShowRunWorkflowDialog(true)}
          size="large"
        >
          <FontAwesomeIcon icon={faPlay} />
          <Typography sx={{ ml: 1 }}>Run Workflow</Typography>
        </Button>
        {runWorkflowDialog}
      </Box>

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
