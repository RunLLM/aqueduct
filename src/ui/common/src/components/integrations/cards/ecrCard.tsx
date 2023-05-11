import React from 'react';

import { ECRConfig, Integration } from '../../../utils/integrations';
import { ResourceCardText } from './text';

type ECRCardProps = {
  integration: Integration;
};

export const ECRCard: React.FC<ECRCardProps> = ({ integration }) => {
  const config = integration.config as ECRConfig;

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
