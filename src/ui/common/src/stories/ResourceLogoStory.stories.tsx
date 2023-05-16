import {ComponentMeta, ComponentStory} from "@storybook/react";
import React from "react";
import IntegrationLogo from "../components/integrations/logo";
import Box from '@mui/material/Box';
import { styled } from '@mui/material/styles';
import {theme} from "../styles/theme/theme";
import SupportedIntegrations from "../utils/SupportedIntegrations";
import {Service} from "../utils/integrations";

// Darken the background so that we can see the component's bounding box.
const BackgroundHighlighter = styled(Box)(() => {
  return {
    backgroundColor: theme.palette.gray[25],
    display: 'inline-flex',
  }
});

const ResourceLogos: React.FC = () => {
  return (
      <Box
          sx={{
            display: 'flex',
            flexWrap: 'wrap',
            alignItems: 'flex-start',
          }}
      >{Object.keys(SupportedIntegrations).map((service) => (
          // TODO: note the padding!
          <BackgroundHighlighter sx={{ml: 2, mt: 2, padding: 1}}>
            <IntegrationLogo service={service as Service} size="large" activated={true} />
          </BackgroundHighlighter>
      ))}
      </Box>
  )

};

const ResourceLogosTemplate: ComponentStory<typeof ResourceLogos> = (args) => (
    <ResourceLogos {...args} />
);

export const ResourceLogosStory = ResourceLogosTemplate.bind({});

export default {
  title: 'Test/ResourceLogos',
  component: ResourceLogos,
  argTypes: {},
} as ComponentMeta<typeof ResourceLogos>;