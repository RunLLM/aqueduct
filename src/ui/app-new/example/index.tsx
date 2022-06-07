import React from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter } from "react-router-dom";
import { Routes, Route, Link } from "react-router-dom";
// import HomePage from './pages/home';
import { Thing } from '../.';
import AboutPage from "./pages/about";
import { HomePage, GettingStartedTutorial } from '@aqueducthq/common';
import { store } from './stores/store';
import { Provider } from 'react-redux';

import { createTheme, ThemeProvider } from '@mui/material/styles';
import { theme } from '@aqueducthq/common/src/styles/theme/theme';

const App = () => {
  const user = {
    name: "Vikram",
    email: "default",
  };

  const muiTheme = createTheme(theme);

  return (
      <ThemeProvider theme={muiTheme}>
        <BrowserRouter>
          <Routes>
            <Route path="/" element={<HomePage user={user} />} />
            <Route path="/about" element={<AboutPage />} />
          </Routes>
        </BrowserRouter>
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
