import React from 'react';

import { FilesystemConfig, Integration } from '../../../utils/integrations';
import { ResourceCardText } from './text';

type FilesystemCardProps = {
  integration: Integration;
};

export const FilesystemCard: React.FC<FilesystemCardProps> = ({
  integration,
}) => {
  const config = integration.config as FilesystemConfig;

  const labels = ['location'];
  const values = [config.location];
  return <ResourceCardText labels={labels} values={values} />;
};

export default FilesystemCard;
