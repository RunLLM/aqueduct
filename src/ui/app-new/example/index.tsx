import * as React from 'react';
import * as ReactDOM from 'react-dom';
import { BrowserRouter } from "react-router-dom";
import { Routes, Route, Link } from "react-router-dom";
import HomePage from './pages/home';
import { Thing } from '../.';
import AboutPage from "./pages/about";

const App = () => {
  return (
      <React.StrictMode>
        <BrowserRouter>
          <Routes>
              <Route path="/" component={HomePage} />
              <Route path="/about" component={AboutPage} />
          </Routes>
        </BrowserRouter>
      </React.StrictMode>
  );
};

ReactDOM.render(<App />, document.getElementById('root'));
