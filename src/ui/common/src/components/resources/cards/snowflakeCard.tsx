import React from 'react';

import { Resource, SnowflakeConfig } from '../../../utils/resources';
import { ResourceCardText } from './text';

type Props = {
  resource: Resource;
  detailedView: boolean;
};

export const SnowflakeCard: React.FC<Props> = ({ resource, detailedView }) => {
  const config = resource.config as SnowflakeConfig;

  let labels = ['Account ID', 'Database', 'User'];
  let values = [config.account_identifier, config.database, config.username];

  if (detailedView) {
    labels = labels.concat(['Warehouse']);
    values = values.concat([config.warehouse]);

    // Only show the Schema field if it was set.
    if (config.schema) {
      labels = labels.concat(['Schema']);
      values = values.concat([config.schema]);
    }

    // Only show the Role field if it was set.
    if (config.role) {
      labels = labels.concat(['Role']);
      values = values.concat([config.role]);
    }
  }

  return <ResourceCardText labels={labels} values={values} />;
};
