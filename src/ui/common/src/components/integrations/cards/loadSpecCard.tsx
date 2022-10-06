import { Typography } from '@mui/material';
import Box from '@mui/material/Box';
import React from 'react';
import { useSelector } from 'react-redux';

import { RootState } from '../../../stores/store';
import { DataPreviewLoadSpec } from '../../../utils/data';
import IntegrationLogo from '../../integrations/logo';

type Props = {
  loadSpecs: DataPreviewLoadSpec[];
};

export const LoadSpecsCard: React.FC<Props> = ({ loadSpecs }) => {
  const integrations = useSelector(
    (state: RootState) => state.integrationsReducer.integrations
  );
  return (
    <Box sx={{ display: 'flex', flexDirection: 'row', alignItems: 'center' }}>
      <Box
        sx={{
          display: 'flex',
          flexDirection: 'row',
          alignItems: 'center',
          marginRight: 1,
          padding: 1,
        }}
      >
        {loadSpecs.map((spec) => {
          const integration = integrations[spec.integration_id];
          return (
            <Box
              marginRight={1}
              key={`load-spec-${spec.integration_id}`}
              display="flex"
              flexDirection="row"
              alignItems="center"
            >
              <IntegrationLogo service={spec.service} activated size="small" />
              {!!integration && (
                <Typography variant="body2" color="gray.700" margin={1}>
                  {integration.name}
                </Typography>
              )}
            </Box>
          );
        })}
      </Box>
    </Box>
  );
};
