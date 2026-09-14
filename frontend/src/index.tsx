import React, { useMemo, useState } from 'react';
import ReactDOM from 'react-dom';
import { Provider } from 'react-redux';
import CssBaseline from '@material-ui/core/CssBaseline';
import { createTheme, ThemeProvider } from '@material-ui/core/styles';
import Routes from './routes';
import store from './redux';
import { init as initSentry } from '@sentry/browser';
import { SENTRY_DSN } from './config';
import 'typeface-roboto';
import './index.css';
// import * as serviceWorker from './serviceWorker';

if (SENTRY_DSN) {
  initSentry({ dsn: SENTRY_DSN });
}

const getInitialMode = () => {
  const savedMode = localStorage.getItem('theme');
  if (savedMode === 'light' || savedMode === 'dark') return savedMode;
  return window.matchMedia('(prefers-color-scheme: dark)').matches
    ? 'dark'
    : 'light';
};

const App = () => {
  const [mode, setMode] = useState<'light' | 'dark'>(getInitialMode);
  const theme = useMemo(() => createTheme({ palette: { type: mode } }), [mode]);
  const toggleMode = () => {
    const nextMode = mode === 'light' ? 'dark' : 'light';
    localStorage.setItem('theme', nextMode);
    setMode(nextMode);
  };

  return (
    <Provider store={store}>
      <ThemeProvider theme={theme}>
        <CssBaseline />
        <Routes mode={mode} onToggleMode={toggleMode} />
      </ThemeProvider>
    </Provider>
  );
};

ReactDOM.render(<App />, document.getElementById('root'));

// If you want your app to work offline and load faster, you can change
// unregister() to register() below. Note this comes with some pitfalls.
// Learn more about service workers: http://bit.ly/CRA-PWA
// serviceWorker.unregister();
