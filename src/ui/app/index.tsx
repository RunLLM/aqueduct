import '@aqueducthq/common/src/styles/globals.css';

import {
    AccountPage,
    ArtifactDetailsPage,
    CheckDetailsPage,
    DataPage,
    ErrorPage,
    HomePage,
    LoginPage,
    MetricDetailsPage,
    OperatorDetailsPage,
    ResourceDetailsPage,
    ResourcesPage,
    WorkflowPage,
    WorkflowsPage,
} from '@aqueducthq/common';
import { UserProfile, useUser } from '@aqueducthq/common';
import { getPathPrefix } from '@aqueducthq/common/src/utils/getPathPrefix';
import React from 'react';
import ReactDOM from 'react-dom/client';
import { Provider } from 'react-redux';
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom';
import { AuthProvider } from '@propelauth/react';

import { store } from './stores/store';

function RequireAuth({ children, user }): { children: JSX.Element; user: UserProfile | undefined } {
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
    const routesContent: React.ReactElement = (
        <Routes>
            <Route
                path={`${pathPrefix ?? '/llm_welcome'}`}
                element={
                    <div>
                        <div>Hello World</div>
                    </div>
                }
            />
            <Route
                path={`${pathPrefix ?? '/'}`}
                element={
                    <RequireAuth user={user}>
                        <HomePage user={user} />{' '}
                    </RequireAuth>
                }
            />
            <Route
                path={`/${pathPrefix}/data`}
                element={
                    <RequireAuth user={user}>
                        <DataPage user={user} />{' '}
                    </RequireAuth>
                }
            />
            <Route
                path={`/${pathPrefix}/resources`}
                element={
                    <RequireAuth user={user}>
                        <ResourcesPage user={user} />{' '}
                    </RequireAuth>
                }
            />
            <Route
                path={`/${pathPrefix}/resource/:id`}
                element={
                    <RequireAuth user={user}>
                        <ResourceDetailsPage user={user} />{' '}
                    </RequireAuth>
                }
            />
            <Route
                path={`/${pathPrefix}/workflows`}
                element={
                    <RequireAuth user={user}>
                        <WorkflowsPage user={user} />{' '}
                    </RequireAuth>
                }
            />
            <Route
                path={`/${pathPrefix}/login`}
                element={user && user.apiKey ? <Navigate to="/" replace /> : <LoginPage />}
            />
            <Route
                path={`/${pathPrefix}/account`}
                element={
                    <RequireAuth user={user}>
                        <AccountPage user={user} />{' '}
                    </RequireAuth>
                }
            />
            <Route
                path={`/${pathPrefix}/workflow/:workflowId`}
                element={
                    <RequireAuth user={user}>
                        <WorkflowPage user={user} />{' '}
                    </RequireAuth>
                }
            />
            <Route
                path={`/${pathPrefix}/workflow/:workflowId/dag/:dagId`}
                element={
                    <RequireAuth user={user}>
                        <WorkflowPage user={user} />{' '}
                    </RequireAuth>
                }
            />
            <Route
                path={`/${pathPrefix}/workflow/:workflowId/result/:dagResultId`}
                element={
                    <RequireAuth user={user}>
                        <WorkflowPage user={user} />{' '}
                    </RequireAuth>
                }
            />
            <Route
                path={`/${pathPrefix}/workflow/:workflowId/result/:dagResultId/operator/:nodeId`}
                element={
                    <RequireAuth user={user}>
                        <OperatorDetailsPage user={user} />{' '}
                    </RequireAuth>
                }
            />
            <Route
                path={`/${pathPrefix}/workflow/:workflowId/result/:dagResultId/artifact/:nodeId`}
                element={
                    <RequireAuth user={user}>
                        <ArtifactDetailsPage user={user} />{' '}
                    </RequireAuth>
                }
            />
            <Route
                path={`/${pathPrefix}/workflow/:workflowId/result/:dagResultId/metric/:nodeId`}
                element={
                    <RequireAuth user={user}>
                        <MetricDetailsPage user={user} />{' '}
                    </RequireAuth>
                }
            />
            <Route
                path={`/${pathPrefix}/workflow/:workflowId/result/:dagResultId/check/:nodeId`}
                element={
                    <RequireAuth user={user}>
                        <CheckDetailsPage user={user} />{' '}
                    </RequireAuth>
                }
            />
            <Route
                path={`/${pathPrefix}/workflow/:workflowId/dag/:dagId/operator/:nodeId`}
                element={
                    <RequireAuth user={user}>
                        <OperatorDetailsPage user={user} />{' '}
                    </RequireAuth>
                }
            />
            <Route
                path={`/${pathPrefix}/workflow/:workflowId/dag/:dagId/artifact/:nodeId`}
                element={
                    <RequireAuth user={user}>
                        <ArtifactDetailsPage user={user} />{' '}
                    </RequireAuth>
                }
            />
            <Route
                path={`/${pathPrefix}/workflow/:workflowId/dag/:dagId/metric/:nodeId`}
                element={
                    <RequireAuth user={user}>
                        <MetricDetailsPage user={user} />{' '}
                    </RequireAuth>
                }
            />
            <Route
                path={`/${pathPrefix}/workflow/:workflowId/dag/:dagId/check/:nodeId`}
                element={
                    <RequireAuth user={user}>
                        <CheckDetailsPage user={user} />{' '}
                    </RequireAuth>
                }
            />
            <Route
                path={`/${pathPrefix}/404`}
                element={
                    user && user.apiKey ? (
                        <RequireAuth user={user}>
                            <ErrorPage user={user} />{' '}
                        </RequireAuth>
                    ) : (
                        <ErrorPage />
                    )
                }
            />
            <Route path="*" element={<Navigate replace to={`/404`} />} />
        </Routes>
    );

    return <BrowserRouter>{routesContent}</BrowserRouter>;
};

const root = ReactDOM.createRoot(document.getElementById('root') as HTMLElement);

// TODO: Put authurl into an environment variable
root.render(
    <AuthProvider authUrl={"https://5729345786.propelauthtest.com"}>
        <Provider store={store}>
            <App />
        </Provider>
    </AuthProvider>,
);
