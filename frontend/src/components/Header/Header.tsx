import React, { useState, useEffect } from 'react';
import Cookies from 'js-cookie';
import AppBar from '@material-ui/core/AppBar';
import IconButton from '@material-ui/core/IconButton';
import Tooltip from '@material-ui/core/Tooltip';
import Toolbar from '@material-ui/core/Toolbar';
import Typography from '@material-ui/core/Typography';
import Brightness4Icon from '@material-ui/icons/Brightness4';
import Brightness7Icon from '@material-ui/icons/Brightness7';
import Search from '../Search';
import useStyles from './useStyles';

interface HeaderProps {
  onSearch: (query: string) => void;
  mode: 'light' | 'dark';
  onToggleMode: () => void;
}

const Header: React.FC<HeaderProps> = ({ onSearch, mode, onToggleMode }) => {
  const [name, setName] = useState('');
  useEffect(() => {
    const name = Cookies.get('user');
    if (name) {
      setName(name);
    }
  }, []);
  const classes = useStyles({});

  return (
    <AppBar position="static">
      <Toolbar>
        <a className={classes.link} href="/go">
          <img src={process.env.PUBLIC_URL + '/logo.svg'} alt="Go URL Logo" className={classes.logo} />
          <Typography variant="h6" color="inherit">
            Go
          </Typography>
        </a>
        <a className={`${classes.link} ${classes.linkSecondary}`} href="/help">
          Help
        </a>
        <div className={classes.grow} />
        <Search onSearch={onSearch} />
        <Tooltip
          title={`Switch to ${mode === 'light' ? 'dark' : 'light'} mode`}
        >
          <IconButton
            color="inherit"
            aria-label="Toggle dark mode"
            data-e2e="theme-toggle"
            onClick={onToggleMode}
          >
            {mode === 'light' ? <Brightness4Icon /> : <Brightness7Icon />}
          </IconButton>
        </Tooltip>

        {name && <span className={classes.name}>{name}</span>}
      </Toolbar>
    </AppBar>
  );
};

export default Header;
