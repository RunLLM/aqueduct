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
  onSelect: (isSelect: boolean, idx: number) => void;
};

export const displayObject = (
  resource_name: string,
  identifier: string,
  update_mode?: UpdateMode | undefined
) => (
  <>
    <Typography variant="body1">
      [{resource_name}] <b>{identifier}</b>
    </Typography>

    {update_mode && (
      <Typography
        style={{
          color: theme.palette.gray[600],
          paddingRight: '8px',
        }}
        variant="body2"
        display="inline"
      >
        Update Mode: {update_mode}
      </Typography>
    )}
  </>
);

const SavedObjectsSelector: React.FC<Props> = ({ objects, onSelect }) => {
  const sortedObjects = [...objects].sort((x, y) => {
    if (x.resource_name !== y.resource_name) {
      return x.resource_name < y.resource_name ? -1 : 1;
    }

    return new Date(x.modified_at) < new Date(y.modified_at) ? -1 : 1;
  });

  return (
    <FormGroup>
      {sortedObjects.map((object, idx) => {
        // Cannot align the checkbox to the top of a multi-line label.
        // Using a weird marginTop workaround.
        return (
          <FormControlLabel
            sx={{ marginTop: '-24px' }}
            key={idx}
            control={
              <Checkbox
                id={idx.toString()}
                onChange={(event) =>
                  onSelect(event.target.checked, parseInt(event.target.id))
                }
              />
            }
            label={
              <Box sx={{ paddingTop: '24px' }}>
                {displayObject(
                  object.resource_name,
                  getSavedObjectIdentifier(object),
                  // update_mode doesn't exist for Googlesheet and will be undefined.
                  object.spec.parameters['update_mode']
                )}
              </Box>
            }
          />
        );
      })}
    </FormGroup>
  );
};

export default SavedObjectsSelector;
