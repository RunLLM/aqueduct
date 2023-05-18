import React from 'react';

import { Service } from '../../utils/resources';
import { SupportedResources } from '../../utils/SupportedResources';

const sizeMap = {
  large: '85px',
  medium: '48px',
  small: '24px',
};

type Props = {
  service: Service;
  size: keyof typeof sizeMap;
  activated: boolean;
};

const ResourceLogo: React.FC<Props> = ({ service, size, activated }) => {
  const logo = SupportedResources[service]?.logo;
  if (!logo) {
    return null;
  }

  const sizePx = sizeMap[size];
  return (
    <img
      src={logo}
      width="100%"
      style={{
        opacity: activated ? 1.0 : 0.3,
        height: sizePx,
        width: sizePx,
        maxWidth: sizePx,
        maxHeight: sizePx,
        objectFit: 'contain',
      }}
    />
  );
};

export default ResourceLogo;
