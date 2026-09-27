import React, { useEffect, useState } from 'react';
import { createRoot } from 'react-dom/client';
import { Provider } from 'react-redux';
import Routes from './routes';
import store from './redux';
import { init as initSentry } from '@sentry/browser';
import { SENTRY_DSN } from './config';
import { Toaster } from './components/ui/toast';
import { TooltipProvider } from './components/ui/tooltip';
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
  useEffect(() => {
    document.documentElement.classList.toggle('dark', mode === 'dark');
  }, [mode]);

  const toggleMode = () => {
    const nextMode = mode === 'light' ? 'dark' : 'light';
    localStorage.setItem('theme', nextMode);
    setMode(nextMode);
  };

  return (
    <Provider store={store}>
      <TooltipProvider>
        <Toaster>
          <Routes mode={mode} onToggleMode={toggleMode} />
        </Toaster>
      </TooltipProvider>
    </Provider>
  );
};

const rootElement = document.getElementById('root');
if (!rootElement) {
  throw new Error('Unable to find the root element');
}
createRoot(rootElement).render(<App />);

// If you want your app to work offline and load faster, you can change
// unregister() to register() below. Note this comes with some pitfalls.
// Learn more about service workers: http://bit.ly/CRA-PWA
// serviceWorker.unregister();
