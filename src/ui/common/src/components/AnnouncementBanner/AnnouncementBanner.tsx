import { faXmark } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Box, Link, Typography } from '@mui/material';
import React, { useEffect, useState } from 'react';
import { useLocation } from 'react-router-dom';

import { theme } from '../../styles/theme/theme';
import UserProfile from '../../utils/auth';
import { apiAddress } from '../hooks/useAqueductConsts';

type AnnouncementBannerProps = {
  onShow: () => void;
  onClose: () => void;
  user: UserProfile;
};

export const AnnouncementBanner: React.FC<AnnouncementBannerProps> = ({
  onShow,
  onClose,
  user,
}) => {
  const location = useLocation();
  const allowedBannerPages = ['/workflows', '/integrations', '/data', '/'];

  // By default do not show banner until we know that we have an announcement to show.
  const [shouldShowAnnouncementBanner, setShouldShowAnnouncementBanner] =
    useState<boolean>(false);
  const [versionNumber, setVersionNumber] = useState<string>('');

  useEffect(() => {
    async function fetchVersionNumber() {
      const res = await fetch(`${apiAddress}/api/version`, {
        method: 'GET',
        headers: { 'api-key': user.apiKey },
      });
      const versionNumberResponse = await res.json();

      const versionBannerDismissed = localStorage.getItem(
        'versionBanner.dismissed'
      );
      let showBanner = false;
      if (versionNumberResponse?.version) {
        const storageResult = localStorage.getItem(
          'versionBanner.lastVersionSeen'
        );
        const versionNumbersStorage = storageResult?.split('.');

        // compare strings to see if the two are equal.
        // if equal, check if banner has been dismissed and return
        if (
          versionNumberResponse.version === storageResult &&
          versionBannerDismissed !== 'true'
        ) {
          showBanner = true;
        } else if (versionNumbersStorage) {
          const versionNumbersResponse =
            versionNumberResponse.version.split('.');
          const majorResponse = parseInt(versionNumbersResponse[0]);
          const minorResponse = parseInt(versionNumbersResponse[1]);
          const patchResponse = parseInt(versionNumbersResponse[2]);

          // compare the two version numbers that we have
          const majorStorage = parseInt(versionNumbersStorage[0]);
          const minorStorage = parseInt(versionNumbersStorage[1]);
          const patchStorage = parseInt(versionNumbersStorage[2]);

          if (
            majorResponse > majorStorage ||
            minorResponse > minorStorage ||
            patchResponse > patchStorage
          ) {
            showBanner = true;
            // Update local storage
            localStorage.setItem(
              'versionBanner.lastVersionSeen',
              versionNumberResponse.version
            );
            // clear dismissed state if user dismissed last banner.
            localStorage.removeItem('versionBanner.dismissed');
          }
          // remember to check if banner has been dismissed.
        } else {
          // newly seen latest version, show banner
          showBanner = true;
          // Update local storage if needed.
          localStorage.setItem(
            'versionBanner.lastVersionSeen',
            versionNumberResponse.version
          );
        }
      }

      setVersionNumber(versionNumberResponse.version);
      setShouldShowAnnouncementBanner(showBanner);
      if (showBanner && onShow) {
        onShow();
      }
    }

    fetchVersionNumber();
  }, [user.apiKey]);

  // Make sure user is on appropriate pages and that the banner should be shown.
  if (
    !shouldShowAnnouncementBanner ||
    allowedBannerPages.indexOf(location.pathname) < 0
  ) {
    if (onClose) {
      onClose();
    }

    return null;
  }

  return (
    <Box
      sx={{
        backgroundColor: theme.palette.gray[100],
        width: '100%',
        height: '32px',
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        position: 'fixed',
        right: 0,
        left: 0,
      }}
    >
      <Box>
        <Typography variant="body1">
          ✨ {versionNumber} has launched!{' '}
          <Link
            href={'https://github.com/aqueducthq/aqueduct/releases'}
            target="_blank"
          >
            Release Notes
          </Link>
        </Typography>
      </Box>
      <Box
        sx={{
          width: '16px',
          fontSize: '16px',
          display: 'flex',
          alignItems: 'center',
          justifySelf: 'space-between',
          position: 'absolute',
          right: '16px',
          cursor: 'pointer',
        }}
      >
        <FontAwesomeIcon
          icon={faXmark}
          onClick={() => {
            if (onClose) {
              onClose();
              localStorage.setItem('versionBanner.dismissed', 'true');
            }
          }}
        />
      </Box>
    </Box>
  );
};

export default AnnouncementBanner;
