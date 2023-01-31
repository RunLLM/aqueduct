import {
  faCircleCheck,
  faCircleXmark,
  faXmark,
} from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
  AlertTitle,
  Checkbox,
  FormGroup,
  List,
  ListItem,
  ListItemIcon,
  ListItemText,
} from '@mui/material';
import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import Divider from '@mui/material/Divider';
import FormControl from '@mui/material/FormControl';
import FormControlLabel, {
  formControlLabelClasses,
} from '@mui/material/FormControlLabel';
import MenuItem from '@mui/material/MenuItem';
import Radio from '@mui/material/Radio';
import RadioGroup from '@mui/material/RadioGroup';
import Select from '@mui/material/Select';
import Snackbar from '@mui/material/Snackbar';
import Switch from '@mui/material/Switch';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import React, { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';

import { handleFetchAllWorkflowSummaries } from '../../reducers/listWorkflowSummaries';
import {
  handleDeleteWorkflow,
  handleListWorkflowSavedObjects,
} from '../../reducers/workflow';
import { AppDispatch, RootState } from '../../stores/store';
import { theme } from '../../styles/theme/theme';
import UserProfile from '../../utils/auth';
import {
  createCronString,
  DayOfWeek,
  deconstructCronString,
  getNextUpdateTime,
  PeriodUnit,
} from '../../utils/cron';
import {
  IntegrationCategories,
  SupportedIntegrations,
} from '../../utils/integrations';
import { UpdateMode } from '../../utils/operators';
import ExecutionStatus, { LoadingStatusEnum } from '../../utils/shared';
import {
  getSavedObjectIdentifier,
  NotificationSettingsMap,
  RetentionPolicy,
  SavedObject,
  WorkflowDag,
  WorkflowUpdateTrigger,
} from '../../utils/workflows';
import { useAqueductConsts } from '../hooks/useAqueductConsts';
import { Button } from '../primitives/Button.styles';
import { LoadingButton } from '../primitives/LoadingButton.styles';
import StorageSelector from './storageSelector';
import TriggerSourceSelector from './triggerSourceSelector';
import WorkflowNotificationSettings from './WorkflowNotificationSettings';

type PeriodicScheduleSelectorProps = {
  cronString: string;
  setSchedule: (string) => void;
};

const PeriodicScheduleSelector: React.FC<PeriodicScheduleSelectorProps> = ({
  cronString,
  setSchedule,
}) => {
  const schedule = deconstructCronString(cronString);

  const [timeUnit, setTimeUnit] = useState(schedule.periodUnit);
  const [minute, setMinute] = useState(schedule.minute);
  const [time, setTime] = useState(schedule.time);
  const [dayOfWeek, setDayOfWeek] = useState(schedule.dayOfWeek);
  const [dayOfMonth, setDayOfMonth] = useState(schedule.dayOfMonth);

  useEffect(() => {
    // Don't try to update the cron schedule if the user enters an invalid
    // input.
    if (
      (timeUnit === PeriodUnit.Hourly && (minute < 0 || minute > 59)) ||
      (timeUnit === PeriodUnit.Monthly && (dayOfMonth < 1 || dayOfMonth > 31))
    ) {
      return;
    }

    setSchedule(
      createCronString({
        periodUnit: timeUnit,
        minute,
        time,
        dayOfWeek,
        dayOfMonth,
      })
    );
  }, [timeUnit, minute, time, dayOfWeek, dayOfMonth, setSchedule]);

  return (
    <Box sx={{ display: 'flex' }}>
      <FormControl size="small" sx={{ mr: 1 }}>
        <Select
          value={timeUnit}
          onChange={(e) => setTimeUnit(e.target.value as PeriodUnit)}
        >
          {Object.values(PeriodUnit).map((option) => (
            <MenuItem key={option} value={option}>
              {option}
            </MenuItem>
          ))}
        </Select>
      </FormControl>

      {timeUnit === 'Monthly' && (
        <TextField
          size="small"
          label="Date"
          sx={{ width: '100px' }}
          type="number"
          value={dayOfMonth}
          onChange={(e) => setDayOfMonth(Number(e.target.value))}
          error={dayOfMonth < 1 || dayOfMonth > 31}
        />
      )}

      {timeUnit === 'Weekly' && (
        <FormControl size="small" sx={{ mx: 1 }}>
          <Select
            value={dayOfWeek}
            onChange={(e) => setDayOfWeek(e.target.value as DayOfWeek)}
          >
            {
              // This is an ugly bit of code. Typescript creates
              // reverse mappings (key->value, value=>key) for
              // numerical enums, so we have to filter out the
              // value->key mappings here before generating the
              // options.
              Object.keys(DayOfWeek)
                .filter((key) => isNaN(Number(key)))
                .map((day) => (
                  <MenuItem key={day} value={DayOfWeek[day]}>
                    {day}
                  </MenuItem>
                ))
            }
          </Select>
        </FormControl>
      )}

      {timeUnit !== 'Hourly' && (
        <TextField
          label="Time"
          sx={{ width: '150px', mx: 1 }}
          size="small"
          type="time"
          value={time}
          onChange={(e) => setTime(e.target.value)}
        />
      )}

      {timeUnit === 'Hourly' && (
        <TextField
          label="Minute"
          sx={{ width: '100px', mx: 1 }}
          size="small"
          type="number"
          value={minute}
          onChange={(e) => setMinute(Number(e.target.value))}
        />
      )}
    </Box>
  );
};

type RetentionPolicyProps = {
  retentionPolicy?: RetentionPolicy;
  setRetentionPolicy: (p?: RetentionPolicy) => void;
};

const RetentionPolicySelector: React.FC<RetentionPolicyProps> = ({
  retentionPolicy,
  setRetentionPolicy,
}) => {
  let value = '';
  let helperText: string = undefined;
  if (!retentionPolicy || retentionPolicy.k_latest_runs <= 0) {
    helperText = 'Aqueduct will store all versions of this workflow.';
  } else {
    value = retentionPolicy.k_latest_runs.toString();
  }

  return (
    <TextField
      size="small"
      label="The number of latest versions to keep. Older versions will be removed."
      fullWidth
      type="number"
      value={value}
      onChange={(e) => {
        const kLatestRuns = parseInt(e.target.value);
        if (kLatestRuns <= 0 || isNaN(kLatestRuns)) {
          // Internal representation of no retention.
          setRetentionPolicy({ k_latest_runs: -1 });
          return;
        }

        setRetentionPolicy({ k_latest_runs: kLatestRuns });
      }}
      helperText={helperText}
    />
  );
};

type WorkflowSettingsProps = {
  user: UserProfile;
  workflowDag: WorkflowDag;
  open: boolean;
  onClose: () => void;
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

const WorkflowSettings: React.FC<WorkflowSettingsProps> = ({
  user,
  workflowDag,
  open,
  onClose,
}) => {
  const { apiAddress } = useAqueductConsts();
  const navigate = useNavigate();

  const dispatch: AppDispatch = useDispatch();

  useEffect(() => {
    dispatch(
      handleListWorkflowSavedObjects({
        apiKey: user.apiKey,
        workflowId: workflowDag.workflow_id,
      })
    );
    dispatch(handleFetchAllWorkflowSummaries({ apiKey: user.apiKey }));
  }, [dispatch, user.apiKey, workflowDag.workflow_id]);

  const savedObjectsResponse = useSelector(
    (state: RootState) => state.workflowReducer.savedObjects
  );

  const savedObjects = savedObjectsResponse.result;
  const savedObjectsStatus = savedObjectsResponse.loadingStatus.loading;

  const [selectedObjects, setSelectedObjects] = useState(
    new Set<SavedObject>()
  );

  const dagResults = useSelector(
    (state: RootState) => state.workflowReducer.dagResults
  );

  const workflows = useSelector(
    (state: RootState) => state.listWorkflowReducer.workflows
  );

  const integrations = useSelector(
    (state: RootState) => state.integrationsReducer.integrations
  );

  const notificationIntegrations = Object.values(integrations).filter(
    (x) =>
      SupportedIntegrations[x.service].category ===
      IntegrationCategories.NOTIFICATION
  );

  const [name, setName] = useState(workflowDag.metadata?.name);
  const [description, setDescription] = useState(
    workflowDag.metadata?.description
  );
  const [triggerType, setTriggerType] = useState(
    workflowDag.metadata.schedule.trigger
  );
  const [schedule, setSchedule] = useState(
    workflowDag.metadata.schedule.cron_schedule
  );
  const [sourceId, setSourceId] = useState(
    workflowDag.metadata?.schedule?.source_id
  );
  const [paused, setPaused] = useState(workflowDag.metadata.schedule.paused);
  const [retentionPolicy, setRetentionPolicy] = useState(
    workflowDag.metadata?.retention_policy
  );
  const [notificationSettingsMap, setNotificationSettingsMap] =
    useState<NotificationSettingsMap>(
      workflowDag.metadata?.notification_settings?.settings ?? {}
    );

  // filter out empty key / values
  const normalizedNotificationSettingsMap = Object.fromEntries(
    Object.entries(notificationSettingsMap).filter(([k, v]) => !!k && !!v)
  );
  const initialSettings = {
    name: workflowDag.metadata?.name,
    description: workflowDag.metadata?.description,
    triggerType: workflowDag.metadata.schedule.trigger,
    schedule: workflowDag.metadata.schedule.cron_schedule,
    paused: workflowDag.metadata.schedule.paused,
    retentionPolicy: workflowDag.metadata?.retention_policy,
    sourceId: workflowDag.metadata?.schedule?.source_id,
    notificationSettingsMap:
      workflowDag.metadata?.notification_settings?.settings ?? {},
  };

  const retentionPolicyUpdated =
    retentionPolicy.k_latest_runs !==
    workflowDag.metadata?.retention_policy?.k_latest_runs;

  const isNotificationSettingsUpdated = IsNotificationSettingsMapUpdated(
    initialSettings.notificationSettingsMap,
    normalizedNotificationSettingsMap
  );

  const settingsChanged =
    name !== workflowDag.metadata?.name || // The workflow name has been changed.
    description !== workflowDag.metadata?.description || // The workflow description has changed.
    triggerType !== workflowDag.metadata.schedule.trigger || // The type of the trigger has changed.
    (triggerType === WorkflowUpdateTrigger.Periodic && // The trigger type is still periodic but the schedule itself has changed.
      schedule !== workflowDag.metadata.schedule.cron_schedule) ||
    (triggerType === WorkflowUpdateTrigger.Cascade && // The trigger type is still cascade but the source has changed.
      sourceId !== workflowDag.metadata?.schedule?.source_id) ||
    paused !== workflowDag.metadata.schedule.paused || // The schedule type is periodic and we've changed the pausedness of the workflow.
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
    workflowDag.metadata?.schedule?.trigger ===
      WorkflowUpdateTrigger.Periodic &&
    !workflowDag.metadata?.schedule?.paused
  ) {
    const nextUpdateTime = getNextUpdateTime(
      workflowDag.metadata?.schedule?.cron_schedule
    );
    nextUpdateComponent = (
      <Box sx={{ fontSize: '10px' }}>
        <Typography variant="body2">
          <strong> Next Workflow Run: </strong>{' '}
          {nextUpdateTime.toDate().toLocaleString()}{' '}
        </Typography>
      </Box>
    );
  }

  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const [
    showSavedObjectDeletionResultsDialog,
    setShowSavedObjectDeletionResultsDialog,
  ] = useState(false);
  const [deleteValidation, setDeleteValidation] = useState('');
  const handleDeleteClicked = (event) => {
    event.preventDefault();
    onClose(); // Close the settings modal.
    setShowDeleteDialog(true);
  };

  // State that controls the Snackbar for an attempted workflow deletion.
  const [deleteMessage, setDeleteMessage] = useState('');
  const [showDeleteMessage, setShowDeleteMessage] = useState(false);

  // State that controls the Snackbar for an attempted workflow settings
  // update.
  const [isUpdating, setIsUpdating] = useState(false);
  const [updateMessage, setUpdateMessage] = useState('');
  const [showUpdateMessage, setShowUpdateMessage] = useState(false);
  const [updateSucceeded, setUpdateSucceeded] = useState(false);

  const savedObjectsDeletionResponse = useSelector(
    (state: RootState) => state.workflowReducer.savedObjectDeletion
  );

  const deleteWorkflowResults = savedObjectsDeletionResponse.result;
  const deleteWorkflowResultsStatus =
    savedObjectsDeletionResponse.loadingStatus.loading;

  let deleteSucceeded = false;
  if (
    deleteWorkflowResultsStatus === LoadingStatusEnum.Succeeded ||
    deleteWorkflowResultsStatus === LoadingStatusEnum.Failed
  ) {
    if (showDeleteDialog) {
      setShowDeleteDialog(false);
    }
    if (deleteWorkflowResultsStatus === LoadingStatusEnum.Succeeded) {
      deleteSucceeded = true;
      if (selectedObjects.size > 0) {
        if (!showSavedObjectDeletionResultsDialog) {
          setShowSavedObjectDeletionResultsDialog(true);
        }
      } else {
        setDeleteMessage(
          'Successfully deleted your workflow. Redirecting you to the workflows page...'
        );
        setShowDeleteMessage(true);
        navigate('/workflows');
      }
    } else if (deleteWorkflowResultsStatus === LoadingStatusEnum.Failed) {
      deleteSucceeded = false;
      setDeleteMessage(
        `We were unable to delete your workflow: ${savedObjectsDeletionResponse.loadingStatus.err}`
      );
      setShowDeleteMessage(true);
      setDeleteValidation('');
    }
  }

  const updateSettings = (event) => {
    event.preventDefault();
    setIsUpdating(true);

    const changes = {
      name: name === workflowDag.metadata?.name ? '' : name,
      description:
        name === workflowDag.metadata?.description ? '' : description,
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
    };

    fetch(`${apiAddress}/api/workflow/${workflowDag.workflow_id}/edit`, {
      method: 'POST',
      headers: {
        'api-key': user.apiKey,
      },
      body: JSON.stringify(changes),
    }).then((res) => {
      res.json().then((body) => {
        setIsUpdating(false);
        if (res.ok) {
          setUpdateSucceeded(true);
          setUpdateMessage('Sucessfully updated your workflow.');
          location.reload(); // Refresh the page to reflect the updated settings.
        } else {
          setUpdateSucceeded(false);
          setUpdateMessage(
            `There was an unexpected error while updating your workflow: ${body.error}`
          );
        }

        setShowUpdateMessage(true);
      });
    });
  };

  const updateSelectedObjects = (event) => {
    if (event.target.checked) {
      setSelectedObjects(
        (prev) => new Set(prev.add(savedObjects[event.target.id][0]))
      );
    } else {
      setSelectedObjects(
        (prev) =>
          new Set(
            Array.from(prev).filter(
              (x) => x !== savedObjects[event.target.id][0]
            )
          )
      );
    }
  };

  const displayObject = (integration, name, sortedObjects) => (
    <>
      <Typography variant="body1">
        [{integration}] <b>{name}</b>
      </Typography>

      {/* Objects saved into S3 are currently expected to have update_mode === UpdateMode.replace */}
      {sortedObjects && (
        <Typography
          style={{
            color: theme.palette.gray[600],
            paddingRight: '8px',
          }}
          variant="body2"
          display="inline"
        >
          Update Mode:{' '}
          {sortedObjects
            .map(
              (object) =>
                `${object.spec.parameters.update_mode || UpdateMode.replace}`
            )
            .join(', ')}
          {sortedObjects.length > 1 && ' (active)'}
        </Typography>
      )}
    </>
  );

  const listSavedObjects = (
    <FormGroup>
      {Object.entries(savedObjects).map(
        ([integrationTableKey, savedObjectsList]) => {
          const sortedObjects = [...savedObjectsList].sort((object) =>
            Date.parse(object.modified_at)
          );

          // Cannot align the checkbox to the top of a multi-line label.
          // Using a weird marginTop workaround.
          return (
            <FormControlLabel
              sx={{ marginTop: '-24px' }}
              key={integrationTableKey}
              control={
                <Checkbox
                  id={integrationTableKey}
                  onChange={updateSelectedObjects}
                />
              }
              label={
                <Box sx={{ paddingTop: '24px' }}>
                  {displayObject(
                    savedObjectsList[0].integration_name,
                    getSavedObjectIdentifier(savedObjectsList[0]),
                    sortedObjects
                  )}
                </Box>
              }
            />
          );
        }
      )}
    </FormGroup>
  );

  const hasSavedObjects = Object.keys(savedObjects).length > 0;

  const deleteDialog = (
    <Dialog
      open={showDeleteDialog}
      onClose={() => {
        setShowDeleteDialog(false);
      }}
      fullWidth
    >
      <DialogTitle>
        <Box sx={{ display: 'flex', alignItems: 'center' }}>
          <Box sx={{ flex: 1 }}>
            <Typography variant="h5">
              {' '}
              {/* We don't use the `name` state here because it will update when the user is mid-changes, which is awkward. */}
              Delete{' '}
              <span style={{ fontFamily: 'Monospace' }}>
                {workflowDag.metadata?.name}
              </span>
              ?{' '}
            </Typography>
          </Box>

          <FontAwesomeIcon
            icon={faXmark}
            onClick={() => setShowDeleteDialog(false)}
            style={{ cursor: 'pointer' }}
          />
        </Box>
      </DialogTitle>

      <DialogContent>
        {hasSavedObjects && (
          <Typography variant="body1">
            The following objects had been saved by{' '}
            <span style={{ fontFamily: 'Monospace' }}>
              {workflowDag.metadata?.name}
            </span>{' '}
            and can be removed when deleting the workflow:
          </Typography>
        )}

        <Box sx={{ my: 2 }}>
          {savedObjectsStatus === LoadingStatusEnum.Succeeded &&
            listSavedObjects}
          {savedObjectsStatus === LoadingStatusEnum.Failed && (
            <Alert severity="error" sx={{ marginTop: 2 }}>
              {`Unable to retrieve list of saved objects. Failed with error: ${savedObjectsResponse.loadingStatus.err}`}
            </Alert>
          )}
        </Box>

        {hasSavedObjects && (
          <Typography variant="body1">
            Deleting workflow{' '}
            <span style={{ fontFamily: 'Monospace' }}>{name}</span> and the
            associated <b>{selectedObjects.size}</b> objects is not reversible.
            Please note that we cannot guarantee this will only delete data
            created by Aqueduct. The workflow will be deleted even if the
            underlying objects are not successfully deleted.
          </Typography>
        )}
        {!hasSavedObjects && (
          <Typography variant="body1">
            Are you sure you want to delete{' '}
            <span style={{ fontFamily: 'Monospace' }}>{name}</span>? This action
            is not reversible.
          </Typography>
        )}

        <Box sx={{ my: 2 }}>
          <Typography variant="body1">
            Type the name of your workflow below to confirm deletion:
          </Typography>
        </Box>

        <TextField
          placeholder={name}
          value={deleteValidation}
          size="small"
          onChange={(e) => setDeleteValidation(e.target.value)}
          fullWidth
        />
      </DialogContent>

      <DialogActions>
        <Button
          variant="outlined"
          color="secondary"
          onClick={() => setShowDeleteDialog(false)}
        >
          Cancel
        </Button>
        <LoadingButton
          variant="contained"
          color="error"
          loading={deleteWorkflowResultsStatus === LoadingStatusEnum.Loading}
          disabled={deleteValidation !== name}
          onClick={(event) => {
            event.preventDefault();
            dispatch(
              handleDeleteWorkflow({
                apiKey: user.apiKey,
                workflowId: workflowDag.workflow_id,
                selectedObjects: selectedObjects,
              })
            );
          }}
        >
          Delete
        </LoadingButton>
      </DialogActions>
    </Dialog>
  );

  let successfullyDeleted = 0;
  let unsuccessfullyDeleted = 0;

  Object.entries(deleteWorkflowResults).map((workflowResults) =>
    workflowResults[1].map((objectResult) => {
      if (objectResult.exec_state.status === ExecutionStatus.Succeeded) {
        successfullyDeleted += 1;
      } else {
        unsuccessfullyDeleted += 1;
      }
    })
  );
  const savedObjectDeletionResultsDialog = (
    <Dialog
      open={showSavedObjectDeletionResultsDialog}
      onClose={() => navigate('/workflows')}
      maxWidth="sm"
      fullWidth
    >
      <DialogTitle>
        <Box sx={{ display: 'flex', alignItems: 'center' }}>
          <Box sx={{ flex: 1 }}>
            <Typography variant="h5">
              {' '}
              {/* We don't use the `name` state here because it will update when the user is mid-changes, which is awkward. */}
              <span style={{ fontFamily: 'Monospace' }}>
                {workflowDag.metadata?.name}
              </span>{' '}
              successfully deleted{' '}
            </Typography>
          </Box>

          <FontAwesomeIcon
            icon={faXmark}
            onClick={() => navigate('/workflows')}
            style={{ cursor: 'pointer' }}
          />
        </Box>
      </DialogTitle>

      <DialogContent>
        <Typography>
          <span style={{ fontFamily: 'Monospace' }}>
            {workflowDag.metadata?.name}
          </span>{' '}
          has been successfully deleted. Here are the results of the saved
          object deletion.
        </Typography>

        <List dense={true}>
          {Object.entries(deleteWorkflowResults)
            .map(([integrationName, objectResults]) =>
              objectResults.map((objectResult) => (
                <>
                  <ListItem key={`${integrationName}-${objectResult.name}`}>
                    <ListItemIcon style={{ minWidth: '30px' }}>
                      {objectResult.exec_state.status ===
                      ExecutionStatus.Succeeded ? (
                        <FontAwesomeIcon
                          icon={faCircleCheck}
                          style={{
                            color: theme.palette.green[500],
                          }}
                        />
                      ) : (
                        <FontAwesomeIcon
                          icon={faCircleXmark}
                          style={{
                            color: theme.palette.red[500],
                          }}
                        />
                      )}
                    </ListItemIcon>
                    <ListItemText
                      primary={displayObject(
                        integrationName,
                        objectResult.name,
                        null
                      )}
                    />
                  </ListItem>
                  {objectResult.exec_state.status ===
                    ExecutionStatus.Failed && (
                    <Alert icon={false} severity="error">
                      <AlertTitle>
                        Failed to delete {objectResult.name}.
                      </AlertTitle>
                      <pre>{objectResult.exec_state.error.context}</pre>
                    </Alert>
                  )}
                </>
              ))
            )
            .flat()}
        </List>

        <Typography>
          <b>Successfully Deleted</b>: {successfullyDeleted}
        </Typography>
        <Typography>
          <b>Unable To Delete</b>: {unsuccessfullyDeleted}
        </Typography>
      </DialogContent>

      <DialogActions>
        <Button variant="contained" onClick={() => navigate('/workflows')}>
          Close
        </Button>
      </DialogActions>
    </Dialog>
  );

  return (
    <>
      <Dialog open={open} onClose={onClose} maxWidth={false}>
        <DialogTitle>
          <Box sx={{ display: 'flex', alignItems: 'center' }}>
            <Box sx={{ flex: 1 }}>
              <Typography variant="h5">
                {' '}
                {/* We don't use the `name` state here because it will update when the user is mid-changes, which is awkward. */}
                <span style={{ fontFamily: 'Monospace' }}>
                  {workflowDag.metadata?.name}
                </span>{' '}
                Settings{' '}
              </Typography>
            </Box>

            <FontAwesomeIcon
              icon={faXmark}
              onClick={() => {
                setName(initialSettings.name);
                setDescription(initialSettings.description);
                setTriggerType(initialSettings.triggerType);
                setSchedule(initialSettings.schedule);
                setSourceId(initialSettings.sourceId);
                setPaused(initialSettings.paused);
                setRetentionPolicy(initialSettings.retentionPolicy);

                // Finally close the dialog
                if (onClose) {
                  onClose();
                }
              }}
              style={{ cursor: 'pointer' }}
            />
          </Box>
        </DialogTitle>

        <DialogContent sx={{ width: '600px' }}>
          <Box sx={{ mb: 2 }}>
            <Box sx={{ mb: 2 }}>
              <Typography sx={{ fontWeight: 'bold' }} component="span">
                ID:
              </Typography>
              <Typography component="span">
                {' '}
                {workflowDag.workflow_id}
              </Typography>
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
            <Typography style={{ fontWeight: 'bold' }}>
              {' '}
              Description{' '}
            </Typography>

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

          {dagResults && dagResults.length > 0 && <StorageSelector />}

          <Box sx={{ my: 2 }}>
            <Typography style={{ fontWeight: 'bold' }}> Schedule </Typography>
            {scheduleSelector}
            {nextUpdateComponent}
          </Box>

          <Box sx={{ my: 2 }}>
            <Typography style={{ fontWeight: 'bold' }}>
              Retention Policy
            </Typography>

            <Box sx={{ my: 1 }}>
              <RetentionPolicySelector
                retentionPolicy={retentionPolicy}
                setRetentionPolicy={setRetentionPolicy}
              />
            </Box>
          </Box>

          {notificationIntegrations.length > 0 && (
            <Box sx={{ my: 2 }}>
              <Typography style={{ fontWeight: 'bold' }}>
                Notifications
              </Typography>

              <WorkflowNotificationSettings
                notificationIntegrations={notificationIntegrations}
                curSettingsMap={notificationSettingsMap}
                onSelect={(id, level) =>
                  setNotificationSettingsMap({
                    ...notificationSettingsMap,
                    [id]: level,
                  })
                }
                onRemove={(id) => {
                  const newSettings = { ...notificationSettingsMap };
                  delete newSettings[id];
                  setNotificationSettingsMap(newSettings);
                }}
              />
            </Box>
          )}

          <LoadingButton
            loading={isUpdating}
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

          <Button
            color="error"
            variant="outlined"
            onClick={handleDeleteClicked}
          >
            Delete Workflow
          </Button>
        </DialogContent>
      </Dialog>
      {deleteDialog}
      {savedObjectDeletionResultsDialog}

      <Snackbar
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
        open={showDeleteMessage}
        onClose={() => setShowDeleteMessage(false)}
        key={'workflowdelete-snackbar'}
        autoHideDuration={6000}
      >
        <Alert
          onClose={() => setShowDeleteMessage(false)}
          severity={deleteSucceeded ? 'success' : 'error'}
          sx={{ width: '100%' }}
        >
          {deleteMessage}
        </Alert>
      </Snackbar>

      <Snackbar
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
        open={showUpdateMessage}
        onClose={() => setShowUpdateMessage(false)}
        key={'settingsupdate-snackbar'}
        autoHideDuration={6000}
      >
        <Alert
          onClose={() => setShowUpdateMessage(false)}
          severity={updateSucceeded ? 'success' : 'error'}
          sx={{ width: '100%' }}
        >
          {updateMessage}
        </Alert>
      </Snackbar>
    </>
  );
};

export default WorkflowSettings;
