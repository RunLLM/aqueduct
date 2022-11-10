import { Box } from '@mui/material';
import React from 'react';

import { PaginatedTable } from '../components/tables/PaginatedTable';
import { Data } from '../utils/data';

export const PaginatedTableStory: React.FC = () => {
  const mockData: Data = {
    schema: {
      fields: [
        { name: 'name', type: 'varchar' },
        { name: 'last_run', type: 'varchar' },
        { name: 'engine', type: 'varchar' },
        { name: 'metrics', type: 'varchar' },
        { name: 'checks', type: 'varchar' },
      ],
      pandas_version: '1.5.1',
    },
    data: [
      {
        name: 'churn',
        last_run: '11/1/2022 2:00PM',
        engine: 'k8s_us_east',
        metrics: 'avg_churn, revenue lost',
        checks: 'min_churn, max_churn',
      },
      {
        name: 'churn',
        last_run: '11/1/2022 2:00PM',
        engine: 'k8s_us_east',
        metrics: 'avg_churn, revenue lost',
        checks: 'min_churn, max_churn',
      },
      {
        name: 'churn',
        last_run: '11/1/2022 2:00PM',
        engine: 'k8s_us_east',
        metrics: 'avg_churn, revenue lost',
        checks: 'min_churn, max_churn',
      },
      {
        name: 'churn',
        last_run: '11/1/2022 2:00PM',
        engine: 'k8s_us_east',
        metrics: 'avg_churn, revenue lost',
        checks: 'min_churn, max_churn',
      },
      {
        name: 'churn',
        last_run: '11/1/2022 2:00PM',
        engine: 'k8s_us_east',
        metrics: 'avg_churn, revenue lost',
        checks: 'min_churn, max_churn',
      },
    ],
  };

  return (
    <Box>
      <PaginatedTable data={mockData} />
    </Box>
  );
};

export default PaginatedTableStory;
