import React from 'react';

import { Resource, S3Config } from '../../../utils/resources';
import { ResourceCardText } from './text';

type Props = {
  resource: Resource;
};

export const S3Card: React.FC<Props> = ({ resource }) => {
  const config = resource.config as S3Config;

  const labels = ['Bucket', 'Region'];
  const values = [config.bucket, config.region];
  if (config.root_dir?.length > 0) {
    labels.push('Root Directory');
    values.push(config.root_dir);
  }

  return <ResourceCardText labels={labels} values={values} />;
};
