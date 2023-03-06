import { faDatabase, faTags } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Box, Typography } from '@mui/material';
import React, { useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { handleGetServerConfig } from '../../handlers/getServerConfig';
import { RootState } from '../../stores/store';
import useUser from '../hooks/useUser';

interface StorageConfigDisplayProps {
  integrationName: string;
}

export const StorageConfigurationDisplay: React.FC<
  StorageConfigDisplayProps
> = ({ integrationName }) => {
  const { user } = useUser();
  const dispatch = useDispatch();
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

  let dataStorageInfo,
    dataStorageText = null;

  if (storageConfig) {
    switch (integrationName) {
      case 'SQLite': {
        if (storageConfig.type === 'SQLite') {
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

        break;
      }
      case 's3': {
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

        dataStorageText = (
          <Typography variant={'body2'}>
            <strong>Storage Type:</strong> {dataStorageInfo}
          </Typography>
        );

        break;
      }
      case 'gcs': {
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

        break;
      }
    }
  }

  return dataStorageText;
};

export default StorageConfigurationDisplay;
