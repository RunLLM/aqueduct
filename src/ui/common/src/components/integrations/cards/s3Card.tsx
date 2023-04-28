import React from 'react';

import { Integration, S3Config } from '../../../utils/integrations';
import { ResourceCardText } from './text';

type Props = {
  integration: Integration;
};

export const S3Card: React.FC<Props> = ({ integration }) => {
  const config = integration.config as S3Config;

  const labels = ['Bucket', 'Region'];
  const values = [config.bucket, config.region];
  if (config.root_dir?.length > 0) {
    labels.push('Root Directory');
    values.push(config.root_dir);
  }

  return <ResourceCardText labels={labels} values={values} />;
};
