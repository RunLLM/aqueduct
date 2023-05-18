import React from 'react';

import { Integration, SQLiteConfig } from '../../../utils/resources';
import { ResourceCardText } from './text';

type SQLiteCardProps = {
  resource: Integration;
};

export const SQLiteCard: React.FC<SQLiteCardProps> = ({ resource }) => {
  const config = resource.config as SQLiteConfig;

  return <ResourceCardText labels={['Database']} values={[config.database]} />;
};

export default SQLiteCard;
