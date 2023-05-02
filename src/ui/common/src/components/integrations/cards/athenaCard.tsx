import React from 'react';

import { AthenaConfig, Integration } from '../../../utils/integrations';
import { ResourceCardText } from './text';

type AthenaCardProps = {
  integration: Integration;
  detailedView: boolean;
};

export const AthenaCard: React.FC<AthenaCardProps> = ({
  integration,
  detailedView,
}) => {
  const config = integration.config as AthenaConfig;

  let labels = ['Database', 'S3 Output Location'];
  let values = [config.database, config.output_location]

  if (detailedView && config.region) {
    labels = labels.concat(['Region']);
    values = values.concat([config.region]);
  }

  if (detailedView && config.config_file_path){
    labels = labels.concat(['Config File Path', "Profile"]);
    values = values.concat([config.config_file_path, config.config_file_profile]);
  }

  return <ResourceCardText labels={labels} values={values} />;
};
