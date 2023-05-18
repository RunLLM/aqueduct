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
import { useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';

import {
  useWorkflowDeletePostMutation,
  useWorkflowEditPostMutation,
  useWorkflowObjectsGetQuery,
  useWorkflowsGetQuery,
} from '../../handlers/AqueductApi';
import {
  DagResponse,
  WorkflowResponse,
} from '../../handlers/responses/workflow';
import { RootState } from '../../stores/store';
import { theme } from '../../styles/theme/theme';
import UserProfile from '../../utils/auth';
import {
  createCronString,
  DayOfWeek,
  deconstructCronString,
  getNextUpdateTime,
  PeriodUnit,
} from '../../utils/cron';
import { IntegrationCategories } from '../../utils/integrations';
import { UpdateMode } from '../../utils/operators';
import ExecutionStatus from '../../utils/shared';
import { SupportedIntegrations } from '../../utils/SupportedIntegrations';
import {
  getSavedObjectIdentifier,
  NotificationSettingsMap,
  RetentionPolicy,
  SavedObject,
  WorkflowUpdateTrigger,
} from '../../utils/workflows';
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

const WorkflowSettings: React.FC<WorkflowSettingsProps> = ({
  user,
  dag,
  workflow,
}) => {
  const navigate = useNavigate();

  const { data: workflows } = useWorkflowsGetQuery({ apiKey: user.apiKey });
  const {
    data: savedObjects,
    error: savedObjectsError,
    isSuccess: savedObjectSuccess,
  } = useWorkflowObjectsGetQuery({
    apiKey: user.apiKey,
    workflowId: workflow.id,
  });
  const [
    deleteWorkflow,
    {
      data: deleteWorkflowResponse,
      isLoading: deleteWorkflowLoading,
      error: deleteWorkflowError,
      isSuccess: deleteWorkflowSuccess,
      reset: resetDeleteWorkflow,
    },
  ] = useWorkflowDeletePostMutation();

  const [editWorkflow, { isLoading: isEditWorkflowLoading }] =
    useWorkflowEditPostMutation();

  const [selectedObjects, setSelectedObjects] = useState(
    new Set<SavedObject>()
  );

  const integrations = useSelector(
    (state: RootState) => state.integrationsReducer.integrations
  );

  const notificationIntegrations = Object.values(integrations).filter(
    (x) =>
      SupportedIntegrations[x.service].category ===
      IntegrationCategories.NOTIFICATION
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

  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const [
    showSavedObjectDeletionResultsDialog,
    setShowSavedObjectDeletionResultsDialog,
  ] = useState(false);
  const [deleteValidation, setDeleteValidation] = useState('');
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

  useEffect(() => {
    if (deleteWorkflowSuccess || !!deleteWorkflowError) {
      if (showDeleteDialog) {
        setShowDeleteDialog(false);
      }

      if (deleteWorkflowSuccess) {
        if (selectedObjects.size > 0) {
          if (!showSavedObjectDeletionResultsDialog) {
            setShowSavedObjectDeletionResultsDialog(true);
          }
        } else {
          navigate('/workflows');
        }
      } else {
        setDeleteValidation('');
      }
    }
  }, [deleteWorkflowSuccess, deleteWorkflowError, navigate]);

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
              <span style={{ fontFamily: 'Monospace' }}>{workflow.name}</span>?{' '}
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
            <span style={{ fontFamily: 'Monospace' }}>{workflow.name}</span> and
            can be removed when deleting the workflow:
          </Typography>
        )}

        <Box sx={{ my: 2 }}>
          {savedObjectSuccess && listSavedObjects}
          {savedObjectsError && (
            <Alert severity="error" sx={{ marginTop: 2 }}>
              {`Unable to retrieve list of saved objects. Failed with error: ${savedObjectsError}`}
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
          loading={deleteWorkflowLoading}
          disabled={deleteValidation !== name}
          onClick={(event) => {
            event.preventDefault();
            const external_delete = {};

            selectedObjects.forEach((object) => {
              if (!external_delete[object.integration_name]) {
                external_delete[object.integration_name] = [];
              }

              external_delete[object.integration_name].push(
                JSON.stringify(object.spec)
              );
            });

            deleteWorkflow({
              apiKey: user.apiKey,
              workflowId: workflow.id,
              force: true,
              external_delete,
            });
          }}
        >
          Delete
        </LoadingButton>
      </DialogActions>
    </Dialog>
  );

  let successfullyDeleted = 0;
  let unsuccessfullyDeleted = 0;

  Object.entries(deleteWorkflowResponse).map((workflowResults) =>
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
      open={deleteWorkflowSuccess}
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
              <span style={{ fontFamily: 'Monospace' }}>{workflow.name}</span>{' '}
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
          <span style={{ fontFamily: 'Monospace' }}>{workflow.name}</span> has
          been successfully deleted. Here are the results of the saved object
          deletion.
        </Typography>

        <List dense={true}>
          {Object.entries(deleteWorkflowResponse)
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

      {notificationIntegrations.length > 0 && (
        <Box sx={{ my: 2 }}>
          <Typography style={{ fontWeight: 'bold' }}>Notifications</Typography>

          <WorkflowNotificationSettings
            notificationIntegrations={notificationIntegrations}
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

      {deleteDialog}
      {savedObjectDeletionResultsDialog}

      <Snackbar
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
        open={!!deleteMessage}
        onClose={() => resetDeleteWorkflow()}
        key={'workflowdelete-snackbar'}
        autoHideDuration={6000}
      >
        <Alert
          onClose={() => resetDeleteWorkflow()}
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
