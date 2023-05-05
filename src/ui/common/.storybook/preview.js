import { theme } from '../src/styles/theme/theme';

import { createTheme, CssBaseline, ThemeProvider } from "@mui/material";

const muiTheme = createTheme(theme);

export const parameters = {
  actions: { argTypesRegex: "^on[A-Z].*" },
  controls: {
    matchers: {
      color: /(background|color)$/i,
      date: /Date$/,
    },
  },
}

// Wrap MUI stories in theme context.
export const withMuiTheme = (Story) => (
  <ThemeProvider theme={muiTheme} >
    <CssBaseline />
    <Story />
  </ThemeProvider>
);

export const decorators = [withMuiTheme];
