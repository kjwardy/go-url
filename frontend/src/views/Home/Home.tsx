import React, { useState, useEffect } from 'react';
import axios from 'axios';
import qs from 'qs';
import { connect } from 'react-redux';
import Paper from '@material-ui/core/Paper';
import Tab from '@material-ui/core/Tab';
import Tabs from '@material-ui/core/Tabs';
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
  };
}

const Home: React.FC<HomeProps> = ({ search, displayFlashError }) => {
  const [querySearchResults, setQuerySearchResults] = useState<any[]>();
  const [popular, setPopular] = useState<any[]>();
  const [activeTab, setActiveTab] = useState(0);
  const classes = useStyles({});
  const match = useRouteMatch<{ query: string }>();
  const location = useLocation();

  useEffect(() => {
    axios.get<any>('/api/popular').then(({ data }) => setPopular(data));
  }, []);

  useEffect(() => {
    const search = location.search;
    const { message } = qs.parse(search.slice(1));
    if (message) {
      displayFlashError(message as string);
    }
  }, [location.search, displayFlashError]);

  useEffect(() => {
    const query = match.params.query;
    if (query) {
      axios
        .get<any>('/api/search', { params: { q: query } })
        .then(({ data }) => setQuerySearchResults(data));
    }
  }, [match.params.query]);

  const created = search.created || [];
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
  const searchResults = search.results || querySearchResults;
  const createdSearchResults = match.params.query
    ? created.filter((result) => result.key.includes(match.params.query))
    : [];

  return (
    <div className={classes.dashboard}>
      <main className={classes.main}>
        {searchResults && (
          <div className={classes.container}>
            <Results
              data={applyUpdates(
                addCreated(searchResults, createdSearchResults),
              )}
              title="Search Results"
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
            <Tab label="Most Popular" />
            <Tab label="Most Wanted" />
            <Tab label="History" />
          </Tabs>
        </Paper>
        {activeTab === 0 && (popular || created.length > 0) && (
          <div className={classes.container}>
            <Results
              data={sortByViews(applyUpdates(addCreated(popular)))}
              title="Most Popular"
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
