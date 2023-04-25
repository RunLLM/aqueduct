import React from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { HomePage, DataPage, IntegrationsPage, IntegrationDetailsPage, WorkflowPage, WorkflowsPage, LoginPage, ErrorPage, AccountPage, OperatorDetailsPage, ArtifactDetailsPage, MetricDetailsPage, CheckDetailsPage } from '@aqueducthq/common';
import { store } from './stores/store';
import { Provider } from 'react-redux';
import { useUser, UserProfile } from '@aqueducthq/common';
import { getPathPrefix } from '@aqueducthq/common/src/utils/getPathPrefix';
import '@aqueducthq/common/src/styles/globals.css';

function RequireAuth({ children, user }): { children: JSX.Element, user: UserProfile | undefined } {
  const pathPrefix = getPathPrefix();

  if (!user || !user.apiKey) {
    return <Navigate to={`${pathPrefix}/login`} replace />;
  }

  return children;
}

const App = () => {
  const { user, loading } = useUser();
  if (loading) {
    return null;
  }

  const pathPrefix = getPathPrefix();
  let routesContent: React.ReactElement;
  routesContent = (
    <Routes>
      <Route path={`${pathPrefix ?? "/"}`} element={<RequireAuth user={user}><HomePage user={user} /> </RequireAuth>} />
      <Route path={`/${pathPrefix}/data`} element={<RequireAuth user={user}><DataPage user={user} /> </RequireAuth>} />
      <Route path={`/${pathPrefix}/integrations`} element={<RequireAuth user={user}><IntegrationsPage user={user} /> </RequireAuth>} />
      <Route path={`/${pathPrefix}/integration/:id`} element={<RequireAuth user={user}><IntegrationDetailsPage user={user} /> </RequireAuth>} />
      <Route path={`/${pathPrefix}/workflows`} element={<RequireAuth user={user}><WorkflowsPage user={user} /> </RequireAuth>} />
      <Route path={`/${pathPrefix}/login`} element={user && user.apiKey ? <Navigate to="/" replace /> : <LoginPage />} />
      <Route path={`/${pathPrefix}/account`} element={<RequireAuth user={user}><AccountPage user={user} /> </RequireAuth>} />
      <Route path={`/${pathPrefix}/workflow/:id`} element={<RequireAuth user={user}><WorkflowPage user={user} /> </RequireAuth>} />
      <Route path={`/${pathPrefix}/workflow/:workflowId/result/:workflowDagResultId/operator/:operatorId`} element={<RequireAuth user={user}><OperatorDetailsPage user={user} /> </RequireAuth>} />
      <Route path={`/${pathPrefix}/workflow/:workflowId/result/:workflowDagResultId/artifact/:artifactId`} element={<RequireAuth user={user}><ArtifactDetailsPage user={user} /> </RequireAuth>} />
      <Route path={`/${pathPrefix}/workflow/:workflowId/result/:workflowDagResultId/metric/:operatorId`} element={<RequireAuth user={user}><MetricDetailsPage user={user} /> </RequireAuth>} />
      <Route path={`/${pathPrefix}/workflow/:workflowId/result/:workflowDagResultId/check/:operatorId`} element={<RequireAuth user={user}><CheckDetailsPage user={user} /> </RequireAuth>} />
      <Route path={`/${pathPrefix}/workflow/:workflowId/dag/:workflowDagId/operator/:operatorId`} element={<RequireAuth user={user}><OperatorDetailsPage user={user} /> </RequireAuth>} />
      <Route path={`/${pathPrefix}/workflow/:workflowId/dag/:workflowDagId/artifact/:artifactId`} element={<RequireAuth user={user}><ArtifactDetailsPage user={user} /> </RequireAuth>} />
      <Route path={`/${pathPrefix}/workflow/:workflowId/dag/:workflowDagId/metric/:operatorId`} element={<RequireAuth user={user}><MetricDetailsPage user={user} /> </RequireAuth>} />
      <Route path={`/${pathPrefix}/workflow/:workflowId/dag/:workflowDagId/check/:operatorId`} element={<RequireAuth user={user}><CheckDetailsPage user={user} /> </RequireAuth>} />
      <Route path={`/${pathPrefix}/404`} element={user && user.apiKey ? <RequireAuth user={user}><ErrorPage user={user} /> </RequireAuth> : <ErrorPage />} />
      <Route path="*" element={<Navigate replace to={`/404`} />} />
    </Routes>
  );

  return (
    <BrowserRouter>{routesContent}</BrowserRouter>
  );
};

const root = ReactDOM.createRoot(
  document.getElementById('root') as HTMLElement
);

root.render(
  <Provider store={store}>
    <App />
  </Provider>
);
