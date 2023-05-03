import React from 'react';

import { Integration, SQLiteConfig } from '../../../utils/integrations';
import { ResourceCardText } from './text';

type SQLiteCardProps = {
  integration: Integration;
};

export const SQLiteCard: React.FC<SQLiteCardProps> = ({ integration }) => {
  const config = integration.config as SQLiteConfig;

  return <ResourceCardText labels={['Database']} values={[config.database]} />;
};

export default SQLiteCard;
