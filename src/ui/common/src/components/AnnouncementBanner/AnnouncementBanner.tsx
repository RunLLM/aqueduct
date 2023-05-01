import { faArrowRight, faXmark } from '@fortawesome/free-solid-svg-icons';
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
  const allowedBannerPages = ['/workflows', '/resources', '/data', '/'];

  // By default do not show banner until we know that we have an announcement to show.
  const [shouldShowAnnouncementBanner, setShouldShowAnnouncementBanner] =
    useState<boolean>(false);
  const [versionNumber, setVersionNumber] = useState<string>('');

  useEffect(() => {
    async function fetchVersionNumber() {
      const pypiRes = await fetch('https://pypi.org/pypi/aqueduct-ml/json', {
        method: 'GET',
      });
      const pyPiResponse = await pypiRes.json();
      const pyPiVersionString = pyPiResponse.info.version;

      const res = await fetch(`${apiAddress}/api/version`, {
        method: 'GET',
        headers: { 'api-key': user.apiKey },
      });
      const aqueductVersionNumberResponse = await res.json();

      const versionBannerDismissed = localStorage.getItem(
        'versionBanner.dismissedVersion'
      );

      let showBanner = false;
      if (aqueductVersionNumberResponse?.version) {
        const pyPiVersionNumbers = pyPiVersionString?.split('.');

        // compare strings to see if the two are equal.
        // if equal, check if banner has been dismissed and return
        const sameVersion =
          aqueductVersionNumberResponse.version === pyPiVersionString;
        const isDismissed = versionBannerDismissed === pyPiVersionString;

        // First check if we should hide the banner.
        if (isDismissed || sameVersion) {
          showBanner = false;
        } else if (pyPiVersionNumbers) {
          const versionNumbersResponse =
            aqueductVersionNumberResponse.version.split('.');
          const majorResponse = parseInt(versionNumbersResponse[0]);
          const minorResponse = parseInt(versionNumbersResponse[1]);
          const patchResponse = parseInt(versionNumbersResponse[2]);

          // compare the two version numbers that we have
          const pyPiMajor = parseInt(pyPiVersionNumbers[0]);
          const pyPiMinor = parseInt(pyPiVersionNumbers[1]);
          const pyPiPatch = parseInt(pyPiVersionNumbers[2]);

          // Finally check if there is in fact a new version and show banner if so.
          if (
            pyPiMajor > majorResponse ||
            pyPiMinor > minorResponse ||
            pyPiPatch > patchResponse
          ) {
            showBanner = true;
          }
        }
      }

      setVersionNumber(pyPiVersionString);
      setShouldShowAnnouncementBanner(showBanner);
      if (showBanner === true && onShow) {
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
        <Typography variant="body1" component={'span'}>
          ✨ Aqueduct v{versionNumber} is out!{' '}
          <Link
            href={'https://github.com/aqueducthq/aqueduct/releases'}
            target="_blank"
          >
            See release notes <FontAwesomeIcon icon={faArrowRight} />
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
              localStorage.setItem(
                'versionBanner.dismissedVersion',
                versionNumber ?? ''
              );
            }
          }}
        />
      </Box>
    </Box>
  );
};

export default AnnouncementBanner;
