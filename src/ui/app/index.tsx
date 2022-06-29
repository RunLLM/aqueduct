import React from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { HomePage, DataPage, IntegrationsPage, IntegrationDetailsPage, WorkflowPage, WorkflowsPage, LoginPage, AccountPage } from '@aqueducthq/common';
import { store } from './stores/store';
import { Provider } from 'react-redux';
import { useUser, UserProfile } from '@aqueducthq/common';
import { createTheme, ThemeProvider } from '@mui/material/styles';
import { theme } from '@aqueducthq/common/src/styles/theme/theme';
import '@aqueducthq/common/src/styles/globals.css';

function RequireAuth({ children, user }): { children: JSX.Element, user: UserProfile | undefined } {
  if (!user || !user.apiKey) {
    return <Navigate to="/login" replace />;
  }

  return children;
}

const App = () => {
  const { user, loading } = useUser();
  if (loading) {
    return null;
  }

  let routesContent: React.ReactElement;
  routesContent = (
    <Routes>
      <Route path={`${ process.env.URL_PREFIX ?? "/" }`} element={<RequireAuth user={user}><HomePage user={user} /> </RequireAuth>} />
      <Route path={`/${process.env.URL_PREFIX}/data`} element={<RequireAuth user={user}><DataPage user={user} /> </RequireAuth>} />
      <Route path={`/${process.env.URL_PREFIX}/integrations`} element={<RequireAuth user={user}><IntegrationsPage user={user} /> </RequireAuth>} />
      <Route path={`/${process.env.URL_PREFIX}/integration/:id`} element={<RequireAuth user={user}><IntegrationDetailsPage user={user} /> </RequireAuth>} />
      <Route path={`/${process.env.URL_PREFIX}/workflows`} element={<RequireAuth user={user}><WorkflowsPage user={user} /> </RequireAuth>} />
      <Route path={`/${process.env.URL_PREFIX}/login`} element={ user && user.apiKey ? <Navigate to="/" replace /> : <LoginPage />} />
      <Route path={`/${process.env.URL_PREFIX}/account`} element={<RequireAuth user={user}><AccountPage user={user} /> </RequireAuth>} />
      <Route path={`/${process.env.URL_PREFIX}/workflow/:id`} element={<RequireAuth user={user}><WorkflowPage user={user} /> </RequireAuth>} />
    </Routes>
  );

  const muiTheme = createTheme(theme);
  return (
      <ThemeProvider theme={muiTheme}>
        <BrowserRouter>{routesContent}</BrowserRouter>
      </ThemeProvider>
  );
};


const root = ReactDOM.createRoot(
  document.getElementById('root') as HTMLElement
);

root.render(
  <React.StrictMode>
    <Provider store={store}>
      <App />
    </Provider>
  </React.StrictMode>
)
