import { faXmark } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
  Alert,
  Box,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  TextField,
  Typography,
} from '@mui/material';
import React, { useEffect, useState } from 'react';

import {
  useWorkflowDeletePostMutation,
  useWorkflowObjectsGetQuery,
} from '../../handlers/AqueductApi';
import { WorkflowResponse } from '../../handlers/responses/workflow';
import UserProfile from '../../utils/auth';
import { SavedObject } from '../../utils/workflows';
import { Button } from '../primitives/Button.styles';
import { LoadingButton } from '../primitives/LoadingButton.styles';
import SavedObjectsSelector from './SavedObjectsSelector';

type Props = {
  user: UserProfile;
  workflow: WorkflowResponse;
  open: boolean;
  onClose: () => void;
};

const DeleteWorkflowDialog: React.FC<Props> = ({
  user,
  workflow,
  open,
  onClose,
}) => {
  const { data: savedObjects, error: savedObjectsError } =
    useWorkflowObjectsGetQuery({
      apiKey: user.apiKey,
      workflowId: workflow.id,
    });

  const [deleteWorkflow, { isLoading: deleteWorkflowLoading }] =
    useWorkflowDeletePostMutation({ fixedCacheKey: `edit-${workflow.id}` });

  const [selectedObjects, setSelectedObjects] = useState(
    new Set<SavedObject>()
  );

  const updateSelectedObjects = (isSelect: boolean, id: string) => {
    if (isSelect) {
      setSelectedObjects((prev) => new Set(prev.add(savedObjects[id][0])));
    } else {
      setSelectedObjects(
        (prev) =>
          new Set(Array.from(prev).filter((x) => x !== savedObjects[id][0]))
      );
    }
  };

  const hasSavedObjects = savedObjects
    ? Object.keys(savedObjects).length > 0
    : false;

  const [deleteValidation, setDeleteValidation] = useState('');

  useEffect(() => setDeleteValidation(''), [open]);

  return (
    <Dialog open={open} onClose={onClose} fullWidth>
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
            onClick={onClose}
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
          {savedObjects?.object_details && (
            <SavedObjectsSelector
              objects={savedObjects.object_details}
              onSelect={updateSelectedObjects}
            />
          )}
          {savedObjectsError && (
            <Alert severity="error" sx={{ marginTop: 2 }}>
              {`Unable to retrieve list of saved objects. Failed with error: ${savedObjectsError}`}
            </Alert>
          )}
        </Box>

        {hasSavedObjects && (
          <Typography variant="body1">
            Deleting workflow{' '}
            <span style={{ fontFamily: 'Monospace' }}>{workflow.name}</span> and
            the associated <b>{selectedObjects.size}</b> objects is not
            reversible. Please note that we cannot guarantee this will only
            delete data created by Aqueduct. The workflow will be deleted even
            if the underlying objects are not successfully deleted.
          </Typography>
        )}
        {!hasSavedObjects && (
          <Typography variant="body1">
            Are you sure you want to delete{' '}
            <span style={{ fontFamily: 'Monospace' }}>{workflow.name}</span>?
            This action is not reversible.
          </Typography>
        )}

        <Box sx={{ my: 2 }}>
          <Typography variant="body1">
            Type the name of your workflow below to confirm deletion:
          </Typography>
        </Box>

        <TextField
          placeholder={workflow.name}
          value={deleteValidation}
          size="small"
          onChange={(e) => setDeleteValidation(e.target.value)}
          fullWidth
        />
      </DialogContent>

      <DialogActions>
        <Button variant="outlined" color="secondary" onClick={onClose}>
          Cancel
        </Button>
        <LoadingButton
          variant="contained"
          color="error"
          loading={deleteWorkflowLoading}
          disabled={deleteValidation !== workflow.name}
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
};

export default DeleteWorkflowDialog;
