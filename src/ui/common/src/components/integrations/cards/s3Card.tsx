import { faDatabase, faTags } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React, { useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { handleGetServerConfig } from '../../../handlers/getServerConfig';
import { RootState } from '../../../stores/store';
import { Integration } from '../../../utils/integrations';
import { S3Config } from '../../../utils/workflows';
import useUser from '../../hooks/useUser';
import StorageConfigurationDisplay from '../StorageConfiguration';

type Props = {
  integration: Integration;
};

export const S3Card: React.FC<Props> = ({ integration }) => {
  const { user } = useUser();
  const dispatch = useDispatch();
  const config = integration.config as S3Config;
  const serverConfig = useSelector(
    (state: RootState) => state.serverConfigReducer
  );

  const storageConfig = serverConfig?.config?.storageConfig;

  useEffect(() => {
    async function fetchServerConfig() {
      if (user) {
        await dispatch(handleGetServerConfig({ apiKey: user.apiKey }));
      }
    }

    fetchServerConfig();
  }, [user]);

  let dataStorageInfo = null;

  if (storageConfig && storageConfig.type === 's3') {
    dataStorageInfo = (
      <Box component="span">
        <Box marginRight="8px" component="span">
          <FontAwesomeIcon icon={faTags} />
        </Box>
        <Box component="span">
          <FontAwesomeIcon icon={faDatabase} />
        </Box>
      </Box>
    );
  } else {
    dataStorageInfo = (
      <Box component="span">
        <FontAwesomeIcon icon={faDatabase} />
      </Box>
    );
  }

  const dataStorageText = (
    <Typography variant={'body2'}>
      <strong>Storage Type:</strong> {dataStorageInfo}
    </Typography>
  );

  return (
    <Box>
      <Typography variant="body2">
        <strong>Bucket: </strong>
        {config.bucket}
      </Typography>
      <Typography variant="body2">
        <strong>Region: </strong>
        {config.region}
      </Typography>
      {dataStorageText}
      <StorageConfigurationDisplay integrationName="s3" />
    </Box>
  );
};
