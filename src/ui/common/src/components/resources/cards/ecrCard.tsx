import React from 'react';

import { ECRConfig, Resource } from '../../../utils/resources';
import { ResourceCardText } from './text';

type ECRCardProps = {
  resource: Resource;
};

export const ECRCard: React.FC<ECRCardProps> = ({ resource }) => {
  const config = resource.config as ECRConfig;

  const labels = [];
  const values = [];

  if (config.region) {
    labels.push('Region');
    values.push(config.region);
  }

  if (config.config_file_path) {
    labels.push('Credential File Path');
    values.push(config.config_file_path);
  }

  if (config.config_file_profile) {
    labels.push('Profile');
    values.push(config.config_file_profile);
  }

  return <ResourceCardText labels={labels} values={values} />;
};

export default ECRCard;
