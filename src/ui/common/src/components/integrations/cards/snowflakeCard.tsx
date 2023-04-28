import React from 'react';

import { Integration, SnowflakeConfig } from '../../../utils/integrations';
import { ResourceCardText } from './text';

type Props = {
  integration: Integration;
  detailedView: boolean;
};

export const SnowflakeCard: React.FC<Props> = ({
  integration,
  detailedView,
}) => {
  const config = integration.config as SnowflakeConfig;

  let labels = ['Account ID', 'Database', 'User'];
  let values = [config.account_identifier, config.database, config.username];

  if (detailedView) {
    labels = labels.concat(['Warehouse', 'Schema', 'Role']);
    values = values.concat([config.warehouse, config.schema, config.role]);
  }

  return <ResourceCardText labels={labels} values={values} />;
};
