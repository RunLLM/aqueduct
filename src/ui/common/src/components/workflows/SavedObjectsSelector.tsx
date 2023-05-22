import {
  Box,
  Checkbox,
  FormControlLabel,
  FormGroup,
  Typography,
} from '@mui/material';
import React from 'react';

import { theme } from '../../styles/theme/theme';
import { UpdateMode } from '../../utils/operators';
import { getSavedObjectIdentifier, SavedObject } from '../../utils/workflows';

type Props = {
  objects: SavedObject[];
  onSelect: (isSelect: boolean, id: string) => void;
};

export const displayObject = (
  integration_name: string,
  object_name: string,
  sortedObjects: SavedObject[]
) => (
  <>
    <Typography variant="body1">
      [{integration_name}] <b>{object_name}</b>
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
              `${object.spec.parameters['update_mode'] || UpdateMode.replace}`
          )
          .join(', ')}
        {sortedObjects.length > 1 && ' (active)'}
      </Typography>
    )}
  </>
);

const SavedObjectsSelector: React.FC<Props> = ({ objects, onSelect }) => {
  const objectsByIntegration: { [integrationName: string]: SavedObject[] } = {};
  objects.forEach((obj) => {
    if (objectsByIntegration[obj.integration_name] === undefined) {
      objectsByIntegration[obj.integration_name] = [];
    }

    objectsByIntegration[obj.integration_name].push(obj);
  });

  return (
    <FormGroup>
      {Object.entries(objectsByIntegration).map(
        ([integrationName, savedObjectList]) => {
          const sortedObjects = [...savedObjectList].sort((object) =>
            Date.parse(object.modified_at)
          );

          // Cannot align the checkbox to the top of a multi-line label.
          // Using a weird marginTop workaround.
          return (
            <FormControlLabel
              sx={{ marginTop: '-24px' }}
              key={integrationName}
              control={
                <Checkbox
                  id={integrationName}
                  onChange={(event) =>
                    onSelect(event.target.checked, event.target.id)
                  }
                />
              }
              label={
                <Box sx={{ paddingTop: '24px' }}>
                  {displayObject(
                    integrationName,
                    getSavedObjectIdentifier(sortedObjects[0]),
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
};

export default SavedObjectsSelector;
