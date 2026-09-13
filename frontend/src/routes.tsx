import {
  Redirect,
  Route,
  BrowserRouter as Router,
  Switch,
} from 'react-router-dom';
import React from 'react';
import Home from './views/Home';
import Layout from './views/Layout';

interface RoutesProps {
  mode: 'light' | 'dark';
  onToggleMode: () => void;
}

const Routes: React.FC<RoutesProps> = ({ mode, onToggleMode }) => (
  <Router basename="/go">
    <Layout mode={mode} onToggleMode={onToggleMode}>
      <Switch>
        <Route exact path="/:query?" component={Home} />
        <Redirect to="/" />
      </Switch>
    </Layout>
  </Router>
);

export default Routes;
