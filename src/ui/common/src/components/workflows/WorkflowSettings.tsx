import { faXmark, faCircleXmark, faCircleCheck, faCircleInfo } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
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
import { useNavigate } from 'react-router-dom';
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
  handleListWorkflowSavedObjects
} from '../../reducers/workflow';
import { SavedObject, WorkflowDag, WorkflowUpdateTrigger } from '../../utils/workflows';
import { useAqueductConsts } from '../hooks/useAqueductConsts';
import { Button } from '../primitives/Button.styles';
import { LoadingButton } from '../primitives/LoadingButton.styles';
import { AppDispatch, RootState } from '../../stores/store';
import { useDispatch, useSelector } from 'react-redux';
import { Checkbox, FormGroup, List, ListItem, ListItemIcon, ListItemText, Tooltip } from '@mui/material';

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
  }, [timeUnit, minute, time, dayOfWeek, dayOfMonth]);

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

type WorkflowSettingsProps = {
  user: UserProfile;
  workflowDag: WorkflowDag;
  open: boolean;
  onClose: () => void;
};

type SavedObjectResult = {
  name: string,
  succeeded: boolean,
}

type DeleteWorkflowResponse = {
  saved_object_deletion_results: { [id: string]: SavedObjectResult[] }
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
    dispatch(handleListWorkflowSavedObjects({ apiKey: user.apiKey, workflowId: workflowDag.workflow_id }));
  }, []);

  const savedObjects = useSelector(
    (state: RootState) => state.workflowReducer.savedObjects
  );

  const [selectedObjects, setSelectedObjects] = useState(new Set<SavedObject>());
  const [deleteWorkflowResults, setDeleteWorkflowResults] = useState({saved_object_deletion_results:{}} as DeleteWorkflowResponse);
  

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
  const [paused, setPaused] = useState(workflowDag.metadata.schedule.paused);

  const settingsChanged =
    name !== workflowDag.metadata?.name || // The workflow name has been changed.
    description !== workflowDag.metadata?.description || // The workflow description has changed.
    triggerType !== workflowDag.metadata.schedule.trigger || // The type of the trigger has changed.
    (triggerType === WorkflowUpdateTrigger.Periodic &&
      schedule !== workflowDag.metadata.schedule.cron_schedule) || // The schedule type is still periodic but the schedule itself has changed.
    paused !== workflowDag.metadata.schedule.paused; // The schedule type is periodic and we've changed the pausedness of the workflow.

  const triggerOptions = [
    { label: 'Update Manually', value: WorkflowUpdateTrigger.Manual },
    { label: 'Update Periodically', value: WorkflowUpdateTrigger.Periodic },
  ];

  const scheduleSelector = (
    <Box sx={{ my: 2 }}>
      <RadioGroup
        onChange={(e) =>
          setTriggerType(e.target.value as WorkflowUpdateTrigger)
        }
        value={triggerType}
        sx={{ width: '200px' }}
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
  const [showSavedObjectDeletionResultsDialog, setShowSavedObjectDeletionResultsDialog] = useState(false);
  const [deleteValidation, setDeleteValidation] = useState('');
  const handleDeleteClicked = (event) => {
    event.preventDefault();
    onClose(); // Close the settings modal.
    setShowDeleteDialog(true);
  };

  // State that controls the Snackbar for an attempted workflow deletion.
  const [isDeleting, setIsDeleting] = useState(false);
  const [deleteMessage, setDeleteMessage] = useState('');
  const [showDeleteMessage, setShowDeleteMessage] = useState(false);
  const [deleteSucceeded, setDeleteSucceeded] = useState(false);

  // State that controls the Snackbar for an attempted workflow settings
  // update.
  const [isUpdating, setIsUpdating] = useState(false);
  const [updateMessage, setUpdateMessage] = useState('');
  const [showUpdateMessage, setShowUpdateMessage] = useState(false);
  const [updateSucceeded, setUpdateSucceeded] = useState(false);

  const handleDeleteWorkflow = (event) => {
    event.preventDefault();
    
    setIsDeleting(true);

    let data = {force: true};
    data['external_delete'] = {};

    selectedObjects.forEach((object) => {
      if (data['external_delete'][object.integration_name]) {
        data['external_delete'][object.integration_name].push(object.object_name)
      } else {
        data['external_delete'][object.integration_name] = [object.object_name]
      }
    });

    fetch(`${apiAddress}/api/workflow/${workflowDag.workflow_id}/delete`, {
      method: 'POST',
      headers: {
        'api-key': user.apiKey,
      },
      body: JSON.stringify(data)
    }).then((res) => {
      res.json().then((body) => {
        setIsDeleting(false);
        setShowDeleteDialog(false);

        if (res.ok) {
          setDeleteSucceeded(true);
          if (selectedObjects.size > 0) {
            setShowSavedObjectDeletionResultsDialog(true);
            setDeleteWorkflowResults(body as DeleteWorkflowResponse);
          } else {
            setDeleteMessage(
              'Successfully deleted your workflow. Redirecting you to the workflows page...'
            );
            setShowDeleteMessage(true);
            navigate('/workflows');
          }
        } else {
          setDeleteSucceeded(false);
          setDeleteMessage(
            `We were unable to delete your workflow: ${body.error}`
          );
          setShowDeleteMessage(true);
          setDeleteValidation('');
        }
      });
    });
  };

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
      },
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
      setSelectedObjects(prev => new Set(prev.add(savedObjects[event.target.id][0])));
    } else {
      setSelectedObjects(prev => new Set( Array.from(prev).filter(x => x !== savedObjects[event.target.id][0])));
    }
  };

  const deleteDialog = (
    <Dialog open={showDeleteDialog} onClose={() => setShowDeleteDialog(false)}>
      <DialogTitle>
        <Typography variant="h5">
          {' '}
          {/* We don't use the `name` state here because it will update when the user is mid-changes, which is awkward. */}
          Delete{' '}
          <span style={{ fontFamily: 'Monospace' }}>
            {workflowDag.metadata?.name}
          </span>
          ?{' '}
        </Typography>
      </DialogTitle>
      
      <DialogContent>
        <Typography variant="body1">
        The following objects had been saved by{' '}
        <span style={{ fontFamily: 'Monospace' }}>
            {workflowDag.metadata?.name}
        </span>
        {' '}and can be removed when deleting the workflow. 
        </Typography>

        <Typography variant="body1">
        Please select the saved objects you wish to delete:
        </Typography>

        <Box sx={{ my: 2 }}>
          <FormGroup>
            {
              Object.entries(savedObjects).map(([integrationTableKey, savedObjectsList]) => (
                  <FormControlLabel 
                  control={<Checkbox 
                    id={integrationTableKey}
                    onChange={updateSelectedObjects}
                  />} 
                  label={
                  <Box>
                    <Typography variant="body1">
                      <b>{savedObjectsList[0].integration_name}</b>: {savedObjectsList[0].object_name}
                    </Typography>

                    <Typography style={{color:theme.palette.gray[600], paddingRight:"8px"}} variant="body2" display="inline">
                      Update Mode: {savedObjectsList.map(object=>object.update_mode).join(", ")}
                    </Typography>

                    <Tooltip title="Multiple update modes have been associated with this object throughout workflow deployments.">
                      <Typography display="inline">
                        <FontAwesomeIcon
                            icon={faCircleInfo}
                            style={{ color:theme.palette.Info }}
                          />
                      </Typography>
                    </Tooltip>
                  </Box>
                  }
                  />
                  )
                )
            }
          </FormGroup>
        </Box>
      
        <Typography variant="body1">
          Are you sure you want to <span style={{color:theme.palette.red[500]}}>delete</span>{' '}
          <span style={{ fontFamily: 'Monospace' }}>{name}</span>? This action
          is not reversible. The workflow and all <b>{selectedObjects.size}</b> selected object(s) {' '}
          <b>regardless of update mode</b> will be <b>completely removed</b>.
        </Typography>

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
          loading={isDeleting}
          disabled={deleteValidation !== name}
          onClick={handleDeleteWorkflow}
        >
          Delete
        </LoadingButton>
      </DialogActions>
    </Dialog>
  );

  const savedObjectDeletionResultsDialog = (
    <Dialog open={showSavedObjectDeletionResultsDialog} onClose={() => setShowSavedObjectDeletionResultsDialog(false)}>
      <DialogTitle>
        <Typography variant="h5">
          Saved Object Deletion Results
        </Typography>
      </DialogTitle>
      
      <DialogContent>
        <List dense={false}>
          {Object.entries(deleteWorkflowResults.saved_object_deletion_results).map(([integrationName, objectResults]) => (
            (objectResults).map((objectResult)=>(
              <ListItem>
                <ListItemIcon>
                  {objectResult.succeeded? 
                  <FontAwesomeIcon
                    icon={faCircleCheck}
                    style={{ color:theme.palette.green[500] }}
                  />
                  :
                  <FontAwesomeIcon
                    icon={faCircleXmark}
                    style={{ color:theme.palette.red[500] }}
                  />
                  }
                </ListItemIcon>
                <ListItemText
                  primary={<><b>{integrationName}</b>: {objectResult.name}</>}
                  secondary={objectResult.succeeded}
                />
              </ListItem>
            ))
          )).flat()}
        </List>
      </DialogContent>

      <DialogActions>
        <Button
          variant="contained"
          onClick={() => navigate('/workflows')}
        >
          Close
        </Button>
      </DialogActions>
    </Dialog>
  );

  return (
    <>
      <Dialog open={open} onClose={onClose}>
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
              onClick={onClose}
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
              <Typography> {workflowDag.workflow_id}</Typography>
            </Box>
          </Box>

          <Box sx={{ my: 2 }}>
            <Typography style={{ fontWeight: 'bold' }}> Name </Typography>

            <TextField
              fullWidth
              value={name}
              onChange={(e) => setName(e.target.value)}
              size="small"
            />
          </Box>

          <Box sx={{ my: 2 }}>
            <Typography style={{ fontWeight: 'bold' }}>
              {' '}
              Description{' '}
            </Typography>

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

          <Box sx={{ my: 2 }}>
            <Typography style={{ fontWeight: 'bold' }}> Schedule </Typography>
            {scheduleSelector}
            {nextUpdateComponent}

            <LoadingButton
              loading={isUpdating}
              onClick={updateSettings}
              sx={{ mt: 1 }}
              color="primary"
              variant="contained"
              disabled={!settingsChanged}
            >
              Save
            </LoadingButton>
          </Box>

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
