import { faTags } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React, { useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { handleGetServerConfig } from '../../../handlers/getServerConfig';
import { RootState } from '../../../stores/store';
import { GCSConfig, Integration } from '../../../utils/integrations';
import useUser from '../../hooks/useUser';
import StorageConfigurationDisplay from '../StorageConfiguration';

type Props = {
  integration: Integration;
};

export const GCSCard: React.FC<Props> = ({ integration }) => {
  const { user } = useUser();
  const dispatch = useDispatch();
  const config = integration.config as GCSConfig;
  const serverConfig = useSelector(
    (state: RootState) => state.serverConfigReducer
  );
  const storageConfig = serverConfig?.config?.storageConfig;

  useEffect(() => {
    async function fetchServerConfig() {
      if (user) {
        dispatch(handleGetServerConfig({ apiKey: user.apiKey }));
      }
    }

    fetchServerConfig();
  }, [user]);

  let dataStorageInfo,
    dataStorageText = null;

  if (storageConfig && storageConfig.type === 'gcs') {
    dataStorageInfo = (
      <Box component="span">
        <FontAwesomeIcon icon={faTags} />
      </Box>
    );

    dataStorageText = (
      <Typography variant={'body2'}>
        <strong>Storage Type:</strong> {dataStorageInfo}
      </Typography>
    );
  }

  return (
    <Box>
      <Typography variant="body2">
        <strong>Bucket: </strong>
        {config.bucket}
      </Typography>
      {dataStorageText}
      <StorageConfigurationDisplay integrationName="gcs" />
    </Box>
  );
};
