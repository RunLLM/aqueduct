import {
  faCircleCheck,
  faCircleXmark,
  faXmark,
} from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
  Alert,
  AlertTitle,
  Box,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  List,
  ListItem,
  ListItemIcon,
  ListItemText,
  Typography,
} from '@mui/material';
import React from 'react';

import { useWorkflowDeletePostMutation } from '../../handlers/AqueductApi';
import { theme } from '../../styles/theme/theme';
import ExecutionStatus from '../../utils/shared';
import { Button } from '../primitives/Button.styles';
import { displayObject } from './SavedObjectsSelector';

type Props = {
  workflowId: string;
  workflowName: string;
  open: boolean;
  onClose: () => void;
};

const SavedObjectDeletionResultDialog: React.FC<Props> = ({
  workflowId,
  workflowName,
  open,
  onClose,
}) => {
  const [_, { data: deleteWorkflowResponse }] = useWorkflowDeletePostMutation({
    fixedCacheKey: `delete-${workflowId}`,
  });

  const deletedObjectsStates =
    deleteWorkflowResponse?.saved_object_deletion_results;

  let successfullyDeleted = 0;
  let unsuccessfullyDeleted = 0;

  if (!deletedObjectsStates || Object.keys(deletedObjectsStates).length === 0) {
    return null;
  }

  Object.entries(deletedObjectsStates).map((workflowResults) =>
    workflowResults[1].map((objectResult) => {
      if (objectResult.exec_state.status === ExecutionStatus.Succeeded) {
        successfullyDeleted += 1;
      } else {
        unsuccessfullyDeleted += 1;
      }
    })
  );

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>
        <Box sx={{ display: 'flex', alignItems: 'center' }}>
          <Box sx={{ flex: 1 }}>
            <Typography variant="h5">
              {' '}
              {/* We don't use the `name` state here because it will update when the user is mid-changes, which is awkward. */}
              <span style={{ fontFamily: 'Monospace' }}>{workflowName}</span>{' '}
              successfully deleted{' '}
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
        <Typography>
          <span style={{ fontFamily: 'Monospace' }}>{workflowName}</span> has
          been successfully deleted. Here are the results of the saved object
          deletion.
        </Typography>

        <List dense={true}>
          {Object.entries(deletedObjectsStates)
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
                        objectResult.name
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
        <Button variant="contained" onClick={onClose}>
          Close
        </Button>
      </DialogActions>
    </Dialog>
  );
};

export default SavedObjectDeletionResultDialog;
