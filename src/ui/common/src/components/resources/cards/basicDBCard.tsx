import React from 'react';

import { Integration } from '../../../utils/resources';
import { ResourceCardText } from './text';

type Props = {
  resource: Integration;
  detailedView: boolean;
};

// Many of the database resources share exactly the same fields: MariaDB, Postgres, MySQL, etc.
type BasicDBConfig = {
  host: string;
  port: string;
  database: string;
  username: string;
};

export const BasicDBCard: React.FC<Props> = ({ resource, detailedView }) => {
  const config = resource.config as BasicDBConfig;

  let labels = ['Host', 'User', 'Database'];
  let values = [config.host, config.username, config.database];

  if (detailedView) {
    labels = labels.concat(['Port']);
    values = values.concat([config.port]);
  }

  return <ResourceCardText labels={labels} values={values} />;
};
