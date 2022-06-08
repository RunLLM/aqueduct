import React, {useEffect, useState} from 'react';
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
import UserProfile from "@aqueducthq/common";

function RequireAuth({ children, user }: { children: JSX.Element, user: UserProfile | undefined }) {
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

  console.log('after loading. user: ', user);
  let routesContent: React.ReactElement;
  routesContent = (
      <Routes>
          <Route path="/" element={<RequireAuth user={user}><HomePage user={user} /></RequireAuth>} />
          <Route path="/data" element={<RequireAuth user={user}><DataPage user={user} /></RequireAuth>} />
          <Route path="/integrations" element={<RequireAuth user={user}><IntegrationsPage user={user} /></RequireAuth>} />
          <Route path="/workflows" element={<RequireAuth user={user}><WorkflowsPage user={user} /></RequireAuth>} />
          <Route path="/login" element={<LoginPage />} />
          <Route path="/workflow/:id" element={<RequireAuth user={user}><WorkflowPage user={user} /></RequireAuth>} />
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
