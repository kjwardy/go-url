import React, { useEffect, useRef, useState } from 'react';
import axios from 'axios';
import { connect } from 'react-redux';
import ClickAwayListener from '@material-ui/core/ClickAwayListener';
import CircularProgress from '@material-ui/core/CircularProgress';
import InputBase from '@material-ui/core/InputBase';
import ListItemText from '@material-ui/core/ListItemText';
import MenuItem from '@material-ui/core/MenuItem';
import MenuList from '@material-ui/core/MenuList';
import Paper from '@material-ui/core/Paper';
import Popper from '@material-ui/core/Popper';
import SearchIcon from '@material-ui/icons/Search';
import Typography from '@material-ui/core/Typography';
import useStyles from './useStyles';

interface SearchResult {
  key: string;
  url: string;
  alias?: string[];
  views: number;
}

interface SearchProps {
  onSearch: (query: string) => void;
  deleted: string[];
}

const SEARCH_DEBOUNCE_MS = 300;

const Search: React.FC<SearchProps> = ({ onSearch, deleted }) => {
  const [query, setQuery] = useState('');
  const [debouncedQuery, setDebouncedQuery] = useState('');
  const [results, setResults] = useState<SearchResult[]>([]);
  const [searchStatus, setSearchStatus] = useState<
    'idle' | 'loading' | 'success' | 'error'
  >('idle');
  const [isFocused, setIsFocused] = useState(false);
  const anchorRef = useRef<HTMLDivElement>(null);
  const classes = useStyles({});
  const trimmedQuery = query.trim();
  const isDebouncing = trimmedQuery !== debouncedQuery;
  const isOpen = isFocused && trimmedQuery.length > 0;
  const visibleResults = results.filter(
    (result) => !deleted.includes(result.key),
  );

  useEffect(() => {
    const timeout = window.setTimeout(
      () => setDebouncedQuery(trimmedQuery),
      SEARCH_DEBOUNCE_MS,
    );
    return () => window.clearTimeout(timeout);
  }, [trimmedQuery]);

  useEffect(() => {
    if (!debouncedQuery) {
      setResults([]);
      setSearchStatus('idle');
      return undefined;
    }

    let isCurrent = true;
    setSearchStatus('loading');
    axios
      .get<SearchResult[]>('/api/search', { params: { q: debouncedQuery } })
      .then(({ data }) => {
        if (!isCurrent) return;
        setResults(data);
        setSearchStatus('success');
      })
      .catch(() => {
        if (!isCurrent) return;
        setSearchStatus('error');
      });

    return () => {
      isCurrent = false;
    };
  }, [debouncedQuery]);

  const searchFor = (searchQuery: string) => {
    const normalizedQuery = searchQuery.trim();
    if (!normalizedQuery) return;
    onSearch(normalizedQuery);
    setIsFocused(false);
  };

  return (
    <ClickAwayListener onClickAway={() => setIsFocused(false)}>
      <div role="search">
        <div ref={anchorRef} className={classes.search}>
          <div className={classes.searchIcon}>
            <SearchIcon />
          </div>
          <InputBase
            aria-label="Search URLs"
            inputProps={{
              'aria-autocomplete': 'list',
              'aria-controls': 'search-suggestions',
              'aria-expanded': isOpen,
              autoComplete: 'off',
            }}
            placeholder="Search…"
            onChange={(event) => setQuery(event.target.value)}
            onFocus={() => setIsFocused(true)}
            onKeyDown={(event) => {
              if (event.key === 'Enter') {
                event.preventDefault();
                searchFor(query);
              }
            }}
            value={query}
            classes={{
              root: classes.inputRoot,
              input: classes.inputInput,
            }}
          />
        </div>
        <Popper
          open={isOpen}
          anchorEl={anchorRef.current}
          placement="bottom-start"
          className={classes.suggestions}
          style={{ width: anchorRef.current?.offsetWidth }}
        >
          <Paper elevation={8}>
            {isDebouncing || searchStatus === 'loading' ? (
              <div className={classes.suggestionState} role="status">
                <CircularProgress size={18} />
                <Typography variant="body2">Searching…</Typography>
              </div>
            ) : searchStatus === 'error' ? (
              <div className={classes.suggestionState} role="status">
                <div>
                  <Typography variant="subtitle2">
                    Search unavailable
                  </Typography>
                  <Typography variant="body2" color="textSecondary">
                    Try again in a moment.
                  </Typography>
                </div>
              </div>
            ) : visibleResults.length === 0 ? (
              <div className={classes.suggestionState} role="status">
                <div>
                  <Typography variant="subtitle2">No results found</Typography>
                  <Typography variant="body2" color="textSecondary">
                    Try a different search term.
                  </Typography>
                </div>
              </div>
            ) : (
              <MenuList
                id="search-suggestions"
                aria-label="Search results"
                role="listbox"
                dense
              >
                {visibleResults.map((result) => (
                  <MenuItem
                    key={result.key}
                    role="option"
                    aria-selected={false}
                    tabIndex={0}
                    onClick={() => searchFor(result.key)}
                  >
                    <ListItemText
                      primary={result.key}
                      secondary={
                        result.alias?.length
                          ? result.alias.join(', ')
                          : result.url
                      }
                      primaryTypographyProps={{ noWrap: true }}
                      secondaryTypographyProps={{ noWrap: true }}
                    />
                  </MenuItem>
                ))}
              </MenuList>
            )}
          </Paper>
        </Popper>
      </div>
    </ClickAwayListener>
  );
};

const mapState = ({ search }) => ({ deleted: search.deleted || [] });

export default connect(mapState)(Search);
