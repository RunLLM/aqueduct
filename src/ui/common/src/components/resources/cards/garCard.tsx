import React from 'react';

import { Resource } from '../../../utils/resources';
import { ResourceCardText } from './text';

type GARCardProps = {
  resource: Resource;
};

export const GARCard: React.FC<GARCardProps> = () => {
  return <ResourceCardText labels={[]} values={[]} />;
};

export default GARCard;
