import React, { useState, useEffect } from 'react';
import axios from 'axios';
import qs from 'qs';
import { connect } from 'react-redux';
import Paper from '@material-ui/core/Paper';
import Tab from '@material-ui/core/Tab';
import Tabs from '@material-ui/core/Tabs';
import ArrowUpwardIcon from '@material-ui/icons/ArrowUpward';
import FastRewindIcon from '@material-ui/icons/FastRewind';
import PriorityHighIcon from '@material-ui/icons/PriorityHigh';
import { useRouteMatch, useLocation } from 'react-router-dom';
import History from '../../components/History';
import Metrics from '../../components/Metrics';
import MostWanted from '../../components/MostWanted';
import Results from '../../components/Results';
import { displayFlashError } from '../../redux/flash/actions';
import useStyles from './useStyles';

interface HomeProps {
  displayFlashError: (message: string) => void;
  search: {
    results?: any[];
    created?: any[];
    updated?: any[];
    deleted?: string[];
  };
}

const Home: React.FC<HomeProps> = ({ search, displayFlashError }) => {
  const [querySearchResults, setQuerySearchResults] = useState<any[]>();
  const [querySearchStatus, setQuerySearchStatus] = useState<
    'loading' | 'loaded' | 'error'
  >('loading');
  const [querySearchRetryCount, setQuerySearchRetryCount] = useState(0);
  const [popular, setPopular] = useState<any[]>();
  const [popularStatus, setPopularStatus] = useState<
    'loading' | 'loaded' | 'error'
  >('loading');
  const [popularRetryCount, setPopularRetryCount] = useState(0);
  const [activeTab, setActiveTab] = useState(0);
  const classes = useStyles({});
  const match = useRouteMatch<{ query: string }>();
  const location = useLocation();
  const query = match.params.query;

  useEffect(() => {
    let cancelled = false;
    setPopularStatus('loading');
    axios
      .get<any[]>('/api/popular')
      .then(({ data }) => {
        if (cancelled) return;
        setPopular(data);
        setPopularStatus('loaded');
      })
      .catch((err) => {
        if (cancelled) return;
        setPopularStatus('error');
        displayFlashError(err.response.data.message || err.response.data);
      });
    return () => {
      cancelled = true;
    };
  }, [displayFlashError, popularRetryCount]);

  useEffect(() => {
    const search = location.search;
    const { message } = qs.parse(search.slice(1));
    if (message) {
      displayFlashError(message as string);
    }
  }, [location.search, displayFlashError]);

  useEffect(() => {
    if (!query) {
      setQuerySearchResults(undefined);
      setQuerySearchStatus('loaded');
      return undefined;
    }

    let cancelled = false;
    setQuerySearchResults([]);
    setQuerySearchStatus('loading');
    axios
      .get<any[]>('/api/search', { params: { q: query } })
      .then(({ data }) => {
        if (cancelled) return;
        setQuerySearchResults(data);
        setQuerySearchStatus('loaded');
      })
      .catch(() => {
        if (cancelled) return;
        setQuerySearchResults([]);
        setQuerySearchStatus('error');
      });
    return () => {
      cancelled = true;
    };
  }, [query, querySearchRetryCount]);

  const created = search.created || [];
  const deleted = search.deleted || [];
  // Place newly created URLs first and remove any duplicate API results
  const addCreated = (results: any[] = [], additions = created) => [
    ...additions,
    ...results.filter(
      (result) =>
        !additions.some((createdUrl) => createdUrl.key === result.key),
    ),
  ];
  const updated = search.updated || [];
  // Replace matching rows while preserving view counts omitted by updates
  const applyUpdates = (results: any[]) =>
    results.map((result) => {
      const updatedUrl = updated.find((url) => url.key === result.key);
      return updatedUrl
        ? { ...result, ...updatedUrl, views: result.views }
        : result;
    });
  const sortByViews = (results: any[]) =>
    [...results].sort((first, second) => second.views - first.views);
  const searchResults = search.results || querySearchResults || [];
  const createdSearchResults = query
    ? created.filter((result) => result.key.includes(query))
    : [];
  const popularResults = sortByViews(
    applyUpdates(addCreated(popular)).filter(
      (result) => !deleted.includes(result.key),
    ),
  );
  const matchingResults = applyUpdates(
    addCreated(searchResults, createdSearchResults),
  ).filter((result) => !deleted.includes(result.key));

  return (
    <div className={classes.dashboard}>
      <main className={classes.main}>
        {query && (
          <div className={classes.container}>
            <Results
              data={matchingResults}
              title="Search Results"
              emptyState={
                querySearchStatus === 'loading'
                  ? {
                      title: 'Searching URLs',
                      description: 'Matching URLs will appear here.',
                      loading: true,
                    }
                  : querySearchStatus === 'error'
                  ? {
                      title: 'Search unavailable',
                      description: 'Try again in a moment.',
                      onRetry: () =>
                        setQuerySearchRetryCount((count) => count + 1),
                    }
                  : {
                      title: 'No matching URLs',
                      description: 'Try a different search term.',
                    }
              }
            />
          </div>
        )}
        <Paper className={classes.tabs}>
          <Tabs
            value={activeTab}
            onChange={(_, value) => setActiveTab(value)}
            indicatorColor="primary"
            textColor="primary"
            aria-label="URL data views"
          >
            <Tab
              label={
                <span className={classes.tabLabel}>
                  <ArrowUpwardIcon
                    className={classes.tabIcon}
                    aria-hidden="true"
                  />
                  Most Popular
                </span>
              }
            />
            <Tab
              label={
                <span className={classes.tabLabel}>
                  <PriorityHighIcon
                    className={classes.tabIcon}
                    aria-hidden="true"
                  />
                  Most Wanted
                </span>
              }
            />
            <Tab
              label={
                <span className={classes.tabLabel}>
                  <FastRewindIcon
                    className={classes.tabIcon}
                    aria-hidden="true"
                  />
                  History
                </span>
              }
            />
          </Tabs>
        </Paper>
        {activeTab === 0 && (
          <div className={classes.container}>
            <Results
              data={popularResults}
              title="Most Popular"
              emptyState={
                popularStatus === 'loading' && popularResults.length === 0
                  ? {
                      title: 'Loading popular URLs',
                      description: 'The most popular URLs will appear here.',
                      loading: true,
                    }
                  : popularStatus === 'error' && popularResults.length === 0
                  ? {
                      title: 'Could not load popular URLs',
                      description:
                        'Try again later. The request failed before the list could be loaded.',
                      onRetry: () => setPopularRetryCount((count) => count + 1),
                    }
                  : {
                      title: 'No popular URLs yet',
                      description: 'Add a URL to get the list started.',
                    }
              }
            />
          </div>
        )}
        {activeTab === 1 && (
          <div className={classes.container}>
            <MostWanted displayFlashError={displayFlashError} />
          </div>
        )}
        {activeTab === 2 && (
          <div className={classes.container}>
            <History displayFlashError={displayFlashError} />
          </div>
        )}
      </main>
      <Metrics displayFlashError={displayFlashError} />
    </div>
  );
};

const mapState = ({ flash, search }) => ({ flash, search });

const mapDispatch = {
  displayFlashError,
};

export default connect(mapState, mapDispatch)(Home);
