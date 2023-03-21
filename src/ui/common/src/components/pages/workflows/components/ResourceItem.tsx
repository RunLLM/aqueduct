import { Box, Tooltip, Typography } from '@mui/material';
import React from 'react';

import { theme } from '../../../../styles/theme/theme';
import { EngineType, EngineTypeToService } from '../../../../utils/engine';
import { ServiceLogos } from '../../../../utils/integrations';

export interface ResourceItemProps {
  resource: string;
  resourceCustomName?: string; // If included, this will override the default name of the engine that we use.
  defaultBackgroundColor?: string; // If included, this will override the default background color if none is specified for the particular engine.
  size?: string; // The font and icon size in pixels.
  collapseName?: boolean; // If true, the name associated with this engine will be in a tooltip. False by default.
}

// These two objects are used when we want to show custom definitions for the ResourceItem. Otherwise,
// we default back to the ServicesLogos imported above.
const ResourceItemBackgrounds = {
  [EngineType.Databricks]: '#FF3B29', // The Databricks logo color.
  [EngineType.K8s]: '#2E6CE6', // The Kubernetes logo color.
  [EngineType.Lambda]: '#FE7E00', // The Lambda logo color.
};

const ResourceItemIcons = {
  [EngineType.Databricks]:
    'https://www.striim.com/wp-content/uploads/2022/06/Databricks-logo-iconwhite.png',
  [EngineType.Spark]:
    'https://cdn.icon-icons.com/icons2/2699/PNG/512/apache_spark_logo_icon_170560.png',
  [EngineType.K8s]:
    'https://cncf-branding.netlify.app/img/projects/kubernetes/icon/white/kubernetes-icon-white.png',
  [EngineType.Lambda]:
    'https://codster.io/wp-content/uploads/2020/08/lambda-y-1-300x300.png',
};

export const ResourceItem: React.FC<ResourceItemProps> = ({
  // The expectation is that we get the internal representation of the engine name,
  // which is all lowercase.
  resource,
  resourceCustomName = undefined,
  defaultBackgroundColor = theme.palette.gray[100],
  size = '16px',
  collapseName = false,
}) => {
  // This is a slightly wonky bit of tech debt, so here's what's happening:
  // If this resource is an engine, then there's a mapping from its internal
  // default name to its proper name. If it's a data system, then there's no
  // mapping. So we check if the resource is a key in the EngineTypeToService
  // map and otherwise fall back to the default name of the resource.
  const resourceName = EngineTypeToService[resource] ?? resource;

  // Same as above, we key the maps in this file on the EngineType because that's how
  // engines are managed, but when we look in ServiceLogos, we use the resourceName
  // which is the "proper" name of the engine.
  let icon: React.ReactElement = (
    <img
      src={ResourceItemIcons[resource] ?? ServiceLogos[resourceName]}
      style={{ marginRight: collapseName ? '0px' : '8px' }}
      width={size}
      height={size}
    />
  );

  if (collapseName) {
    icon = (
      <Tooltip title={resourceCustomName ?? resourceName} arrow>
        {icon}
      </Tooltip>
    );
  }

  return (
    <Box
      display="flex"
      alignItems="center"
      px={1}
      py="4px"
      width="fit-content"
      sx={{
        borderRadius: '8px',
        backgroundColor:
          ResourceItemBackgrounds[resource] ?? defaultBackgroundColor,
        color: ResourceItemBackgrounds[resource] ? 'white' : 'black',
      }}
      textOverflow="ellipsis"
    >
      {icon}
      {!collapseName && (
        <Typography overflow="hidden" fontSize={size} variant="body1">
          {resourceCustomName ?? resourceName}
        </Typography>
      )}
    </Box>
  );
};

export default ResourceItem;
