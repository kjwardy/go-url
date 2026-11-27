import React, { useState, useEffect, useCallback } from 'react';
import { connect } from 'react-redux';
import { useHistory, useLocation } from 'react-router-dom';
import AddIcon from '@material-ui/icons/Add';
import Fab from '@material-ui/core/Fab';
import Tooltip from '@material-ui/core/Tooltip';
import EditModal from '../../components/EditModal';
import Header from '../../components/Header';
import Alert from '../../components/Alert';
import { Variant } from '../../components/Alert/SnackbarContentWrapper';
import { SERVICE_VERSION } from '../../config';
import useStyles from './useStyles';

interface LayoutProps {
  mode: 'light' | 'dark';
  onToggleMode: () => void;
  flash: {
    message: string;
    variant: Variant;
  };
}

const Layout: React.FC<LayoutProps> = ({
  children,
  flash,
  mode,
  onToggleMode,
}) => {
  const [addOpen, setAddOpen] = useState(false);
  const [query, onSearch] = useState<string>();
  const classes = useStyles({});
  const hideAdd = useCallback(() => setAddOpen(false), []);
  const history = useHistory();
  const location = useLocation();
  // Pre-populate field if not found
  const urlQuery =
    location.search.includes('message=') &&
    decodeURIComponent(location.pathname.slice(1));

  useEffect(() => {
    if (query === undefined) return;
    history.push(`/${encodeURIComponent(query)}`);
  }, [query, history]);

  return (
    <div className={classes.root}>
      {flash.message && (
        <Alert variant={flash.variant} message={flash.message} />
      )}
      {addOpen && (
        <EditModal onClose={hideAdd} urlKey={urlQuery || undefined} />
      )}
      <Header onSearch={onSearch} mode={mode} onToggleMode={onToggleMode} />
      {children}
      <footer className={classes.footer}>
        <span>go-url {SERVICE_VERSION}</span>
        <span aria-hidden="true">|</span>
        <a
          className={classes.githubLink}
          href="https://github.com/kjwardy/go-url"
          target="_blank"
          rel="noopener noreferrer"
        >
          <svg
            aria-hidden="true"
            className={classes.githubIcon}
            viewBox="0 0 16 16"
          >
            <path
              fill="currentColor"
              d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38v-1.33c-2.23.49-2.7-.95-2.7-.95-.36-.92-.89-1.16-.89-1.16-.73-.5.06-.49.06-.49.8.06 1.22.82 1.22.82.71 1.21 1.87.86 2.33.66.07-.52.28-.86.51-1.06-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82A7.65 7.65 0 0 1 8 4.52c.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48v1.54c0 .21.15.46.55.38A8.01 8.01 0 0 0 16 8c0-4.42-3.58-8-8-8Z"
            />
          </svg>
          GitHub
        </a>
      </footer>
      <Tooltip title="Add New URL">
        <Fab
          data-e2e="add-button"
          color="secondary"
          aria-label="Add"
          className={classes.button}
          onClick={() => setAddOpen(true)}
        >
          <AddIcon />
        </Fab>
      </Tooltip>
    </div>
  );
};

const mapState = ({ flash }) => ({ flash });
export default connect(mapState)(Layout);
