import React from 'react';

import { Resource, SQLiteConfig } from '../../../utils/resources';
import { ResourceCardText } from './text';

type SQLiteCardProps = {
  resource: Resource;
};

export const SQLiteCard: React.FC<SQLiteCardProps> = ({ resource }) => {
  const config = resource.config as SQLiteConfig;

  return <ResourceCardText labels={['Database']} values={[config.database]} />;
};

export default SQLiteCard;
