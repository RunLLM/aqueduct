import { Box } from '@mui/material';
import React from 'react';

import {Card} from "../components/layouts/card";
import {IntegrationCard} from "../components/integrations/cards/card";
import {Integration, PostgresConfig} from "../utils/integrations";

export const ResourceCardStory: React.FC = () => {
  const integrations: Integration[] = [
    {
      id: '1',
      service: "Postgres",
      name: "Postgres Resource",
      config: {
        host: "aam19861.us-east-2.amazonaws.com",
        port: "5432",
        database: "prod",
        username: "prod-pg-aq",
        password: "this is a password",
      } as PostgresConfig,
      createdAt: Date.now() / 1000,
    },
    {
      id: '2',
      service: "Snowflake",
      
    }
  ];

  // Is missing the <Link> component that encapsulates the <Card> component.
  return (
      <Box
          sx={{
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'flex-start',
          }}
      >
        {[...integrations]
            .sort((a, b) => (a.createdAt < b.createdAt ? 1 : -1))
            .map((integration, idx) => {
              return (
                  <Box key={idx} sx={{ width: '90%', maxWidth: '1000px' }}>
                      <Card sx={{ my: 2 }}>
                        <IntegrationCard integration={integration} />
                      </Card>
                  </Box>
              );
            })}
      </Box>
  );
};

export default ResourceCardStory;
