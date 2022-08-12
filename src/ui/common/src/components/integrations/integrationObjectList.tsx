import { faRefresh } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
  Alert,
  Autocomplete,
  Link,
  TextField,
  Typography,
} from '@mui/material';
import Box from '@mui/material/Box';
import React, { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import {
  handleListIntegrationObjects,
  handleLoadIntegrationObject,
  objectKeyFn,
} from '../../reducers/integration';
import { AppDispatch, RootState } from '../../stores/store';
import UserProfile from '../../utils/auth';
import { Integration } from '../../utils/integrations';
import { isLoading } from '../../utils/shared';
import IntegrationObjectPreview from './integrationObjectPreview';

type Props = {
  user: UserProfile;
  integration: Integration;
  onUploadCsv?: () => void;
};

const IntegrationObjectList: React.FC<Props> = ({ user, integration }) => {
  const listObjectNamesState = useSelector(
    (state: RootState) => state.integrationReducer.objectNames
  );
  const objectsState = useSelector(
    (state: RootState) => state.integrationReducer.objects
  );
  const dispatch: AppDispatch = useDispatch();
  const [selectedObject, setSelectedObject] = useState<string>('');
  const objectKey = objectKeyFn(selectedObject);
  const objectState = objectsState[objectKey];
  const hasObject = !!selectedObject && !!objectState;

  useEffect(() => {
    dispatch(
      handleLoadIntegrationObject({
        apiKey: user.apiKey,
        integrationId: integration.id,
        object: selectedObject,
      })
    );
  }, [selectedObject]);

  if (integration.service === 'S3') {
    return (
      <Alert severity="warning" sx={{ width: '80%', mt: 4 }}>
        <>
          We currently do not support listing data in an S3 bucket. But
          don&apos;t worry&mdash;we&apos;re working on adding this feature! If
          you have questions, comments or would like to learn more about what
          we&apos;re building, please{' '}
        </>
        <Link href="mailto:hello@aqueducthq.com">reach out</Link>
        <>, </>
        <Link href="https://join.slack.com/t/aqueductusers/shared_invite/zt-11hby91cx-cpmgfK0qfXqEYXv25hqD6A">
          join our Slack channel
        </Link>
        <>, or </>
        <Link href="https://github.com/aqueducthq/aqueduct/issues/new">
          start a conversation on GitHub channel
        </Link>
        <>.</>
      </Alert>
    );
  }

  const listObjectNamesLoading = isLoading(listObjectNamesState.status);
  return (
    <Box sx={{ mt: 4 }}>
      <Typography variant="h4" gutterBottom component="div">
        Preview
      </Typography>
      <Box
        display="flex"
        flexDirection="row"
        alignContent="center"
        alignItems="center"
      >
        <Autocomplete
          disablePortal
          value={selectedObject}
          sx={{
            verticalAlign: 'middle',
            display: 'inline-block',
            width: '35ch',
          }}
          onChange={(_, val: string) => setSelectedObject(val)}
          options={listObjectNamesState.names}
          loading={listObjectNamesLoading}
          renderInput={(params) => (
            <TextField
              {...params}
              label="Base Table"
              InputProps={{
                ...params.InputProps,
                endAdornment: (
                  <React.Fragment>
                    {params.InputProps.endAdornment}
                  </React.Fragment>
                ),
              }}
            />
          )}
        />
        <FontAwesomeIcon
          className={listObjectNamesLoading ? 'fa-spin' : ''}
          style={{
            marginLeft: '15px',
            fontSize: '2em',
            verticalAlign: 'middle',
            display: 'inline-block',
            color: listObjectNamesLoading ? 'grey' : 'black',
            cursor: listObjectNamesLoading ? 'default' : 'pointer',
          }}
          icon={faRefresh}
          onClick={() => {
            if (!listObjectNamesLoading) {
              dispatch(
                handleListIntegrationObjects({
                  apiKey: user.apiKey,
                  integrationId: integration.id,
                  forceLoad: true,
                })
              );
            }
          }}
        />
      </Box>

      {hasObject && (
        <IntegrationObjectPreview
          objectName={selectedObject}
          object={objectState}
        />
      )}
    </Box>
  );
};

export default IntegrationObjectList;
