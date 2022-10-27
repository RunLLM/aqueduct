// import styles to load faster.
// see here for more info: 
// https://storybook.js.org/blog/material-ui-in-storybook/
import '@fontsource/roboto/300.css';
import '@fontsource/roboto/400.css';
import '@fontsource/roboto/500.css';
import '@fontsource/roboto/700.css';
import '@fontsource/material-icons';
//import '@fontsource/roboto/300.css'
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
