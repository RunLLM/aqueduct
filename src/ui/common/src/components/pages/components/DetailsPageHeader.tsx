import { faCircleCheck } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import Box from '@mui/material/Box';
import Link from '@mui/material/Link';
import Typography from '@mui/material/Typography';
import React from 'react';

type DetailsPageHeaderProps = {
  name: string;
  createdAt?: string;
  sourceLocation?: string;
};

export const DetailsPageHeader: React.FC<DetailsPageHeaderProps> = ({
  name,
  // TODO: add these back once we have support for getting createdAt and sourceLocation.
  createdAt,
  sourceLocation,
}) => {
  return (
    <Box width="100%" display="flex" alignItems="center">
      <FontAwesomeIcon
        height="24px"
        width="24px"
        style={{ marginRight: '8px' }}
        icon={faCircleCheck}
        color={'green'}
      />
      <Typography variant="h4" component="div">
        {name}
      </Typography>
      {createdAt && (
        <Typography marginTop="4px" variant="caption" component="div">
          Created: {createdAt}
        </Typography>
      )}

      {sourceLocation && (
        <Typography variant="caption" component="div">
          Source: <Link>{sourceLocation}</Link>
        </Typography>
      )}
    </Box>
  );
};

export default DetailsPageHeader;
