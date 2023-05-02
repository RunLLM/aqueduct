import React from 'react';

import { DatabricksConfig, Integration } from '../../../utils/integrations';
import { ResourceCardText } from './text';

type DatabricksCardProps = {
  integration: Integration;
  detailedView: boolean;
};

export const DatabricksCard: React.FC<DatabricksCardProps> = ({
  integration,
  detailedView,
}) => {
  const config = integration.config as DatabricksConfig;

  let labels = ['Workspace', 'S3 Instance Profile ARN'];
  let values = [config.workspace_url, config.s3_instance_profile_arn];


  if (detailedView && config.instance_pool_id) {
    labels = labels.concat(['Instance Pool ID']);
    values = values.concat([config.instance_pool_id]);
  }

  return <ResourceCardText labels={labels} values={values} />;
};
