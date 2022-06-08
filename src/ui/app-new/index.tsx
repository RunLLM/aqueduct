import React from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter, Routes, Route, Link, Navigate, useLocation } from "react-router-dom";
import { Thing } from '../.';
import { HomePage, DataPage, IntegrationsPage, WorkflowPage, WorkflowsPage, LoginPage } from '@aqueducthq/common';
import { store } from './stores/store';
import { Provider } from 'react-redux';
import { useUser } from '@aqueducthq/common';

import { createTheme, ThemeProvider } from '@mui/material/styles';
import { theme } from '@aqueducthq/common/src/styles/theme/theme';
import '@aqueducthq/common/src/styles/globals.css';

const App = () => {
  const { user, loading, success } = useUser();
  if (loading) {
    return null;
  }

  let routesContent: React.ReactElement;
  if (!success) {
    routesContent = (
      <Routes>
        <Route path="/" element={<Navigate to="/login" />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/data" element={<Navigate to="/login" />} />
        <Route path="/integrations" element={<Navigate to="/login" />} />
        <Route path="/workflows" element={<Navigate to="/login" />} />
        <Route path="/workflow/:id" element={<Navigate to="/login" />} />
      </Routes>
    );
  } else {
    routesContent = (
      <Routes>
        <Route path="/" element={<HomePage user={user} />} />
        <Route path="/data" element={<DataPage user={user} />} />
        <Route path="/integrations" element={<IntegrationsPage user={user} />} />
        <Route path="/workflows" element={<WorkflowsPage user={user} />} />
        <Route path="/login" element={<Navigate to="/" />} />
        <Route path="/workflow/:id" element={<WorkflowPage user={user} />} />
      </Routes>
    );
  }

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
