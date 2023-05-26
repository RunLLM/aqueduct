import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Divider from '@mui/material/Divider';
import FormControlLabel, {
  formControlLabelClasses,
} from '@mui/material/FormControlLabel';
import Radio from '@mui/material/Radio';
import RadioGroup from '@mui/material/RadioGroup';
import Snackbar from '@mui/material/Snackbar';
import Switch from '@mui/material/Switch';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import React, { useEffect, useState } from 'react';
import { useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';

import {
  useWorkflowDeletePostMutation,
  useWorkflowEditPostMutation,
  useWorkflowsGetQuery,
} from '../../handlers/AqueductApi';
import {
  DagResponse,
  WorkflowResponse,
} from '../../handlers/responses/workflow';
import { RootState } from '../../stores/store';
import UserProfile from '../../utils/auth';
import { getNextUpdateTime } from '../../utils/cron';
import { ResourceCategories } from '../../utils/resources';
import { SupportedResources } from '../../utils/SupportedResources';
import {
  NotificationSettingsMap,
  WorkflowUpdateTrigger,
} from '../../utils/workflows';
import { Button } from '../primitives/Button.styles';
import { LoadingButton } from '../primitives/LoadingButton.styles';
import DeleteWorkflowDialog from './DeleteWorkflowDialog';
import PeriodicScheduleSelector from './PeriodicScheduleSelector';
import RetentionPolicySelector from './RetentionPolicySelector';
import SavedObjectDeletionResultDialog from './SavedObjectDeletionResultDialog';
import StorageSelector from './storageSelector';
import TriggerSourceSelector from './triggerSourceSelector';
import WorkflowNotificationSettings from './WorkflowNotificationSettings';

type Props = {
  user: UserProfile;
  workflow: WorkflowResponse;
  dag: DagResponse;
};

// Returns whether `updated` is different from `existing`.
function IsNotificationSettingsMapUpdated(
  curSettingsMap: NotificationSettingsMap,
  newSettingsMap: NotificationSettingsMap
): boolean {
  // Starting here, both `curSettings` and `newSettings` should be non-empty.
  if (
    Object.keys(curSettingsMap).length !== Object.keys(newSettingsMap).length
  ) {
    return true;
  }

  // both should have the same key size. Check k-v match
  let updated = false;
  Object.entries(curSettingsMap).forEach(([k, v]) => {
    if (newSettingsMap[k] !== v) {
      updated = true;
    }
  });
  return updated;
}

const WorkflowSettings: React.FC<Props> = ({ user, dag, workflow }) => {
  const navigate = useNavigate();

  const { data: workflows, refetch: refetchWorkflows } = useWorkflowsGetQuery({
    apiKey: user.apiKey,
  });
  const [
    {},
    {
      data: deleteWorkflowResponse,
      error: deleteWorkflowError,
      isSuccess: deleteWorkflowSuccess,
      reset: resetDeleteWorkflow,
    },
  ] = useWorkflowDeletePostMutation({ fixedCacheKey: `delete-${workflow.id}` });

  const [editWorkflow, { isLoading: isEditWorkflowLoading }] =
    useWorkflowEditPostMutation({ fixedCacheKey: `edit-${workflow.id}` });

  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const [
    showSavedObjectDeletionResultsDialog,
    setShowSavedObjectDeletionResultsDialog,
  ] = useState(false);

  const handleDeleteClicked = (event) => {
    event.preventDefault();
    setShowDeleteDialog(true);
  };

  // State that controls the Snackbar for an attempted workflow deletion.
  const deleteMessage = deleteWorkflowSuccess
    ? 'Successfully deleted your workflow. Redirecting you to the workflows page...'
    : deleteWorkflowError
    ? `We were unable to delete your workflow: ${deleteWorkflowError}`
    : '';

  const handleDeleteMessageClose = () => {
    if (deleteWorkflowSuccess) {
      refetchWorkflows();
      navigate('/workflows');
      navigate(0); // force refresh the page.
      return;
    }

    resetDeleteWorkflow();
  };

  useEffect(() => {
    if (deleteWorkflowSuccess || !!deleteWorkflowError) {
      setShowDeleteDialog(false);

      if (deleteWorkflowSuccess) {
        if (
          Object.keys(deleteWorkflowResponse.saved_object_deletion_results)
            .length > 0
        ) {
          setShowSavedObjectDeletionResultsDialog(true);
        }
      }
    }
  }, [deleteWorkflowSuccess, deleteWorkflowError, navigate]);

  const resources = useSelector(
    (state: RootState) => state.resourcesReducer.resources
  );

  const notificationResources = Object.values(resources).filter(
    (x) =>
      SupportedResources[x.service].category === ResourceCategories.NOTIFICATION
  );

  const [name, setName] = useState(workflow.name);
  const [description, setDescription] = useState(workflow?.description);
  const [triggerType, setTriggerType] = useState(workflow.schedule.trigger);
  const [schedule, setSchedule] = useState(workflow.schedule.cron_schedule);
  const [sourceId, setSourceId] = useState(workflow.schedule?.source_id);
  const [paused, setPaused] = useState(workflow.schedule.paused);
  const [retentionPolicy, setRetentionPolicy] = useState(
    workflow.retention_policy
  );
  const [notificationSettingsMap, setNotificationSettingsMap] =
    useState<NotificationSettingsMap>(
      workflow.notification_settings?.settings ?? {}
    );

  // filter out empty key / values
  const normalizedNotificationSettingsMap = Object.fromEntries(
    Object.entries(notificationSettingsMap).filter(([k, v]) => !!k && !!v)
  );
  const initialSettings = {
    name: workflow.name,
    description: workflow.description,
    triggerType: workflow.schedule.trigger,
    schedule: workflow.schedule.cron_schedule,
    paused: workflow.schedule.paused,
    retentionPolicy: workflow.retention_policy,
    sourceId: workflow.schedule?.source_id,
    notificationSettingsMap: workflow.notification_settings?.settings ?? {},
  };

  const retentionPolicyUpdated =
    retentionPolicy.k_latest_runs !== workflow.retention_policy?.k_latest_runs;

  const isNotificationSettingsUpdated = IsNotificationSettingsMapUpdated(
    initialSettings.notificationSettingsMap,
    normalizedNotificationSettingsMap
  );

  const settingsChanged =
    name !== workflow.name || // The workflow name has been changed.
    description !== workflow.description || // The workflow description has changed.
    triggerType !== workflow.schedule.trigger || // The type of the trigger has changed.
    (triggerType === WorkflowUpdateTrigger.Periodic && // The trigger type is still periodic but the schedule itself has changed.
      schedule !== workflow.schedule.cron_schedule) ||
    (triggerType === WorkflowUpdateTrigger.Cascade && // The trigger type is still cascade but the source has changed.
      sourceId !== workflow.schedule?.source_id) ||
    paused !== workflow.schedule.paused || // The schedule type is periodic and we've changed the pausedness of the workflow.
    retentionPolicyUpdated ||
    isNotificationSettingsUpdated; // retention policy has changed.

  const triggerOptions = [
    { label: 'Update Manually', value: WorkflowUpdateTrigger.Manual },
    { label: 'Update Periodically', value: WorkflowUpdateTrigger.Periodic },
    {
      label: 'Update After Completion Of',
      value: WorkflowUpdateTrigger.Cascade,
    },
  ];

  const scheduleSelector = (
    <Box sx={{ my: 1 }}>
      <RadioGroup
        onChange={(e) =>
          setTriggerType(e.target.value as WorkflowUpdateTrigger)
        }
        value={triggerType}
        sx={{ width: '250px' }}
      >
        {triggerOptions.map(({ label, value }) => {
          return (
            <FormControlLabel
              value={value}
              label={label}
              control={<Radio size="small" disableRipple />}
              key={value}
              sx={{
                [`& .${formControlLabelClasses.label}`]: { fontSize: '14px' },
              }}
            />
          );
        })}
      </RadioGroup>

      {triggerType === WorkflowUpdateTrigger.Periodic && (
        <>
          <PeriodicScheduleSelector
            setSchedule={setSchedule}
            cronString={schedule}
          />
          <FormControlLabel
            sx={{ mt: 1, ml: 0 }}
            label="Pause Workflow"
            control={
              <Switch
                size="small"
                onChange={() => setPaused(!paused)}
                checked={paused}
              />
            }
          />
        </>
      )}

      {triggerType === WorkflowUpdateTrigger.Cascade && (
        <TriggerSourceSelector
          sourceId={sourceId}
          setSourceId={setSourceId}
          workflows={workflows}
        />
      )}
    </Box>
  );

  let nextUpdateComponent;
  if (
    workflow.schedule?.trigger === WorkflowUpdateTrigger.Periodic &&
    !workflow.schedule?.paused
  ) {
    const nextUpdateTime = getNextUpdateTime(workflow.schedule?.cron_schedule);
    nextUpdateComponent = (
      <Box sx={{ fontSize: '10px' }}>
        <Typography variant="body2">
          <strong> Next Workflow Run: </strong>{' '}
          {nextUpdateTime.toDate().toLocaleString()}{' '}
        </Typography>
      </Box>
    );
  }

  const updateSettings = (event) => {
    event.preventDefault();

    editWorkflow({
      apiKey: user.apiKey,
      workflowId: workflow.id,
      name: name === workflow.name ? '' : name,
      description: name === workflow.description ? '' : description,
      schedule: {
        trigger: triggerType, // We always set the trigger type to be safe because it's stored as a single JSON blob.
        cron_schedule:
          triggerType === WorkflowUpdateTrigger.Periodic ? schedule : '', // Always set the schedule if the update type is periodic.
        paused, // Set whatever value of paused was set, which will be the previous value if it's not modified.
        source_id:
          triggerType === WorkflowUpdateTrigger.Cascade
            ? sourceId
            : '00000000-0000-0000-0000-000000000000',
      },
      retention_policy: retentionPolicyUpdated ? retentionPolicy : undefined,
      notification_settings: isNotificationSettingsUpdated
        ? { settings: normalizedNotificationSettingsMap }
        : undefined,
    });
  };

  return (
    <>
      <Box sx={{ my: 2 }}>
        <Box sx={{ mb: 2 }}>
          <Typography sx={{ fontWeight: 'bold' }} component="span">
            ID:
          </Typography>
          <Typography component="span"> {workflow.id}</Typography>
        </Box>
      </Box>

      <Box sx={{ my: 2 }}>
        <Typography style={{ fontWeight: 'bold' }}> Name </Typography>

        <Box sx={{ my: 1 }}>
          <TextField
            fullWidth
            value={name}
            onChange={(e) => setName(e.target.value)}
            size="small"
          />
        </Box>
      </Box>

      <Box sx={{ my: 2 }}>
        <Typography style={{ fontWeight: 'bold' }}> Description </Typography>

        <Box sx={{ my: 1 }}>
          <TextField
            fullWidth
            placeholder="Your description goes here."
            value={description}
            multiline
            rows={4}
            size="small"
            onChange={(e) => setDescription(e.target.value)}
          />
        </Box>
      </Box>

      <StorageSelector dag={dag} />

      <Box sx={{ my: 2 }}>
        <Typography style={{ fontWeight: 'bold' }}> Schedule </Typography>
        {scheduleSelector}
        {nextUpdateComponent}
      </Box>

      <Box sx={{ my: 2 }}>
        <Typography style={{ fontWeight: 'bold' }}>Retention Policy</Typography>

        <Box sx={{ my: 1 }}>
          <RetentionPolicySelector
            retentionPolicy={retentionPolicy}
            setRetentionPolicy={setRetentionPolicy}
          />
        </Box>
      </Box>

      {notificationResources.length > 0 && (
        <Box sx={{ my: 2 }}>
          <Typography style={{ fontWeight: 'bold' }}>Notifications</Typography>

          <WorkflowNotificationSettings
            notificationResources={notificationResources}
            curSettingsMap={notificationSettingsMap}
            onSelect={(id, level, replacingID) => {
              const newSettings = { ...notificationSettingsMap };
              newSettings[id] = level;
              if (replacingID) {
                delete newSettings[replacingID];
              }

              setNotificationSettingsMap(newSettings);
            }}
            onRemove={(id) => {
              const newSettings = { ...notificationSettingsMap };
              delete newSettings[id];
              setNotificationSettingsMap(newSettings);
            }}
          />
        </Box>
      )}

      <Button
        color="info"
        variant="outlined"
        sx={{ marginRight: 2 }}
        disabled={!settingsChanged}
        onClick={() => {
          setName(initialSettings.name);
          setDescription(initialSettings.description);
          setTriggerType(initialSettings.triggerType);
          setSchedule(initialSettings.schedule);
          setSourceId(initialSettings.sourceId);
          setPaused(initialSettings.paused);
          setRetentionPolicy(initialSettings.retentionPolicy);
          setNotificationSettingsMap(initialSettings.notificationSettingsMap);
        }}
      >
        Discard Changes
      </Button>

      <LoadingButton
        loading={isEditWorkflowLoading}
        onClick={updateSettings}
        sx={{ my: 1 }}
        color="primary"
        variant="contained"
        disabled={!settingsChanged}
      >
        Save
      </LoadingButton>

      <Divider />

      <Box sx={{ my: 2 }}>
        <Typography variant="h6"> Danger Zone </Typography>
      </Box>

      <Button color="error" variant="outlined" onClick={handleDeleteClicked}>
        Delete Workflow
      </Button>

      <DeleteWorkflowDialog
        open={showDeleteDialog}
        onClose={() => setShowDeleteDialog(false)}
        workflow={workflow}
        user={user}
      />
      <SavedObjectDeletionResultDialog
        open={showSavedObjectDeletionResultsDialog}
        onClose={handleDeleteMessageClose}
        workflowName={workflow.name}
        workflowId={workflow.id}
      />

      <Snackbar
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
        open={!!deleteMessage}
        onClose={handleDeleteMessageClose}
        key={'workflowdelete-snackbar'}
        autoHideDuration={6000}
      >
        <Alert
          onClose={handleDeleteMessageClose}
          severity={deleteWorkflowSuccess ? 'success' : 'error'}
          sx={{ width: '100%' }}
        >
          {deleteMessage}
        </Alert>
      </Snackbar>
    </>
  );
};

export default WorkflowSettings;
