import React from 'react';

import { FilesystemConfig, Resource } from '../../../utils/resources';
import { ResourceCardText } from './text';

type FilesystemCardProps = {
  resource: Resource;
};

export const FilesystemCard: React.FC<FilesystemCardProps> = ({ resource }) => {
  const config = resource.config as FilesystemConfig;

  const labels = ['location'];
  const values = [config.location];
  return <ResourceCardText labels={labels} values={values} />;
};

export default FilesystemCard;
