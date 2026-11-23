import {
  Redirect,
  Route,
  BrowserRouter as Router,
  Switch,
} from 'react-router-dom';
import React from 'react';
import Home from './views/Home';
import Layout from './views/Layout';
import Login from './views/Login';

interface RoutesProps {
  mode: 'light' | 'dark';
  onToggleMode: () => void;
}

const Routes: React.FC<RoutesProps> = ({ mode, onToggleMode }) => (
  <Router basename="/go">
    <Switch>
      <Route exact path="/login" component={Login} />
      <Route>
        <Layout mode={mode} onToggleMode={onToggleMode}>
          <Switch>
            <Route exact path="/:query?" component={Home} />
            <Redirect to="/" />
          </Switch>
        </Layout>
      </Route>
    </Switch>
  </Router>
);

export default Routes;
