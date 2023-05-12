import { Box } from '@mui/material';
import { ComponentMeta } from '@storybook/react';
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
        name: 'predict_churn',
        last_run: '11/1/2022 2:00PM',
        engine: 'k8s_us_east',
        metrics: 'avg_churn, revenue lost',
        checks: 'min_churn, max_churn',
      },
      {
        name: 'mpg_prediction',
        last_run: '11/1/2022 2:00PM',
        engine: 'k8s_us_east',
        metrics: 'avg_mpg',
        checks: 'avg_mpg_reasonable',
      },
      {
        name: 'housing_price_prediction',
        last_run: '11/1/2022 2:00PM',
        engine: 'k8s_us_east',
        metrics: 'price_increase, max_price_by_zip_code',
        checks: 'price_increase_reasonable',
      },
      {
        name: 'venture_capital_prediction',
        last_run: '11/1/2022 2:00PM',
        engine: 'k8s_us_east',
        metrics: 'revenue, profit, debt',
        checks: 'reasonable_debt',
      },
      {
        name: 'world_cup_prediction',
        last_run: '11/1/2022 2:00PM',
        engine: 'k8s_us_east',
        metrics: 'avg_goals',
        checks: 'avg_goals_reasonable',
      },
    ],
  };

  return (
    <Box>
      <PaginatedTable data={mockData} />
    </Box>
  );
};

export default {
  title: 'Components/PaginatedTable',
  component: PaginatedTableStory,
  argTypes: {},
} as ComponentMeta<typeof PaginatedTableStory>;
