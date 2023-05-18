import React from 'react';

import { AthenaConfig, Integration } from '../../../utils/resources';
import { ResourceCardText } from './text';

type AthenaCardProps = {
  resource: Integration;
  detailedView: boolean;
};

export const AthenaCard: React.FC<AthenaCardProps> = ({
  resource,
  detailedView,
}) => {
  const config = resource.config as AthenaConfig;

  let labels = ['Database', 'S3 Output Location'];
  let values = [config.database, config.output_location];

  if (detailedView && config.region) {
    labels = labels.concat(['Region']);
    values = values.concat([config.region]);
  }

  if (detailedView && config.config_file_path) {
    labels = labels.concat(['Config File Path', 'Profile']);
    values = values.concat([
      config.config_file_path,
      config.config_file_profile,
    ]);
  }

  return <ResourceCardText labels={labels} values={values} />;
};
