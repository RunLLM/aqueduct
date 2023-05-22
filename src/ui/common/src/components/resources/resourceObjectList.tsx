import {
  Accordion,
  AccordionDetails,
  AccordionSummary,
  Alert,
  CircularProgress,
  Link,
  Typography,
} from '@mui/material';
import Box from '@mui/material/Box';
import React, { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { handleLoadResourceObject, objectKeyFn } from '../../reducers/resource';
import { AppDispatch, RootState } from '../../stores/store';
import { theme } from '../../styles/theme/theme';
import UserProfile from '../../utils/auth';
import { Resource } from '../../utils/resources';
import { isLoading } from '../../utils/shared';
import ResourceObjectPreview from './resourceObjectPreview';

type Props = {
  user: UserProfile;
  resource: Resource;
  onUploadCsv?: () => void;
};

const DefaultTableListLimit = 5;

const ResourceObjectList: React.FC<Props> = ({ user, resource }) => {
  const listObjectNamesState = useSelector(
    (state: RootState) => state.resourceReducer.objectNames
  );
  const objectsState = useSelector(
    (state: RootState) => state.resourceReducer.objects
  );

  const [limitTableList, setLimitTableList] = useState(true);

  const dispatch: AppDispatch = useDispatch();
  const [selectedObject, setSelectedObject] = useState<string>('');
  const [openPanel, setOpenPanel] = useState<number>(-1);

  useEffect(() => {
    dispatch(
      handleLoadResourceObject({
        apiKey: user.apiKey,
        resourceId: resource.id,
        object: selectedObject,
      })
    );
  }, [dispatch, resource.id, selectedObject, user.apiKey]);

  if (resource.service === 'Kubernetes' || resource.service === 'Lambda') {
    return null;
  }

  if (resource.service === 'S3') {
    return (
      <Alert severity="warning" sx={{ width: '100%', mt: 2 }}>
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

  const handleChange = (idx: number) => {
    if (openPanel === idx) {
      // Close the panel we previously opened.
      setOpenPanel(-1);
      setSelectedObject('');
    } else {
      // Open a new panel.
      setOpenPanel(idx);
      setSelectedObject(listObjectNamesState.names[idx]);
    }
  };

  const tablesList = [];
  const tableListLimit: number =
    limitTableList && DefaultTableListLimit < listObjectNamesState.names.length
      ? DefaultTableListLimit
      : listObjectNamesState.names.length;
  for (let i = 0; i < tableListLimit; i++) {
    const element = (
      <Accordion
        expanded={openPanel === i}
        sx={{ width: '100%' }}
        key={i}
        onChange={() => handleChange(i)}
      >
        <AccordionSummary sx={{ backgroundColor: theme.palette.gray[25] }}>
          {' '}
          {listObjectNamesState.names[i]}{' '}
        </AccordionSummary>
        <AccordionDetails>
          <ResourceObjectPreview
            objectName={selectedObject}
            object={objectsState[objectKeyFn(selectedObject)]}
          />
        </AccordionDetails>
      </Accordion>
    );
    tablesList.push(element);
  }

  const tablesDisplay: React.ReactNode = (
    <Box width="900px">
      {tablesList}

      {listObjectNamesState.names.length > DefaultTableListLimit && (
        <Typography
          variant="body2"
          sx={{
            textDecoration: 'underline',
            color: theme.palette.blue[400],
            cursor: 'pointer',
            mt: 1,
          }}
          onClick={() => setLimitTableList(!limitTableList)}
        >
          See {limitTableList ? 'more' : 'fewer'} tables...
        </Typography>
      )}
    </Box>
  );

  const listObjectNamesLoading = isLoading(listObjectNamesState.status);
  return (
    <Box sx={{ mt: 4 }}>
      <Typography variant="h5" gutterBottom component="div">
        Data
      </Typography>

      <Typography variant="body2" sx={{ mb: 1 }}>
        These are the tables stored in {resource.name}. You can click into any
        of the tables below to see a preview of the data.
      </Typography>

      <Box
        display="flex"
        flexDirection="row"
        alignContent="center"
        alignItems="center"
      >
        {listObjectNamesLoading ? <CircularProgress /> : <>{tablesDisplay}</>}
      </Box>
    </Box>
  );
};

export default ResourceObjectList;
