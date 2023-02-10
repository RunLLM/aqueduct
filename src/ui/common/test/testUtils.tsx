import { createTheme, CssBaseline, ThemeProvider } from '@mui/material';
import { render } from '@testing-library/react';
import React from 'react';

import { theme } from '../src/styles/theme/theme';
import { UserProfile } from '../src/utils/auth';

const muiTheme = createTheme(theme);
export const AqueductThemeProvider = ({ children }: any) => {
  return (
    <ThemeProvider theme={muiTheme}>
      <CssBaseline />
      {children}
    </ThemeProvider>
  );
};

// TODO: Add other contexts here like our router / history context
const renderWithProviders = (ui: React.ReactElement) => {
  return render(<AqueductThemeProvider>{ui}</AqueductThemeProvider>);
};

export const mockUser: UserProfile = {
  name: 'Test User',
  email: 'testuser@aqueducthq.com',
  // auth0 related prop
  sub: 'testsub',
  apiKey: 'UI_TEST_API_KEY',
  // come up with some time.
  updated_at: '1/1/1 10:00pm PST',
  email_verified: true,
  nickname: 'TestNickname',
  // TODO: add base64 encoded picture string. (this came from auth0)
  // picture:
  given_name: 'Test',
  family_name: 'User',
};

/*
export type UserProfile = {
  name?: string;
  email?: string;
  sub?: string;
  apiKey?: string;
  updated_at?: string;
  email_verified?: boolean;
  nickname?: string;
  picture?: string;
  organizationId?: string;
  userRole?: string;
  given_name?: string;
  family_name?: string;
};
*/

export { screen } from '@testing-library/react';
export { renderWithProviders as render };
