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
  const theme = useMemo(
    () =>
      createTheme({
        palette: {
          type: mode,
          primary: { main: mode === 'light' ? '#4054b2' : '#8ea2ff' },
          secondary: { main: '#00a58e' },
          background: {
            default: mode === 'light' ? '#f4f6fb' : '#11141b',
            paper: mode === 'light' ? '#ffffff' : '#1b202a',
          },
        },
        shape: { borderRadius: 10 },
        typography: {
          fontFamily: 'Roboto, sans-serif',
          h4: { fontWeight: 600 },
          h6: { fontWeight: 600 },
          button: { fontWeight: 600, textTransform: 'none' },
        },
        overrides: {
          MuiAppBar: {
            colorPrimary: {
              color: '#ffffff',
              backgroundColor: mode === 'light' ? '#33469b' : '#1b202a',
            },
          },
          MuiPaper: {
            rounded: { borderRadius: 12 },
          },
          MuiTableCell: {
            head: {
              fontWeight: 600,
              color: mode === 'light' ? '#4d5669' : '#c5cada',
              backgroundColor: mode === 'light' ? '#f7f8fc' : '#222833',
            },
            root: {
              borderBottomColor:
                mode === 'light'
                  ? 'rgba(34, 45, 72, 0.08)'
                  : 'rgba(255, 255, 255, 0.08)',
            },
          },
          MuiTableRow: {
            root: {
              '&:hover': {
                backgroundColor:
                  mode === 'light'
                    ? 'rgba(64, 84, 178, 0.04)'
                    : 'rgba(142, 162, 255, 0.06)',
              },
            },
          },
          MuiDialogTitle: {
            root: { padding: '24px 24px 8px' },
          },
          MuiDialogActions: {
            root: { padding: '8px 24px 20px' },
          },
          MuiButton: {
            root: { borderRadius: 8, paddingLeft: 16, paddingRight: 16 },
          },
          MuiFab: {
            root: { boxShadow: '0 8px 24px rgba(20, 29, 60, 0.24)' },
          },
        },
      }),
    [mode],
  );
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
