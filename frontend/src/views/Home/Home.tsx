import React, { useState, useEffect } from 'react';
import axios from 'axios';
import qs from 'qs';
import { connect } from 'react-redux';
import { useRouteMatch, useLocation } from 'react-router-dom';
import Results from '../../components/Results';
import { displayFlashError } from '../../redux/flash/actions';
import useStyles from './useStyles';

interface HomeProps {
  displayFlashError: (message: string) => void;
  search: {
    results?: any[];
    created?: any[];
  };
}

const Home: React.FC<HomeProps> = ({ search, displayFlashError }) => {
  const [querySearchResults, setQuerySearchResults] = useState<any[]>();
  const [popular, setPopular] = useState<any[]>();
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
  const searchResults = search.results || querySearchResults;
  const createdSearchResults = match.params.query
    ? created.filter((result) => result.key.includes(match.params.query))
    : [];

  return (
    <div>
      {searchResults && (
        <div className={classes.container}>
          <Results
            data={addCreated(searchResults, createdSearchResults)}
            title="Search Results"
          />
        </div>
      )}
      {(popular || created.length > 0) && (
        <div className={classes.container}>
          <Results data={addCreated(popular)} title="Most Popular" />
        </div>
      )}
    </div>
  );
};

const mapState = ({ flash, search }) => ({ flash, search });

const mapDispatch = {
  displayFlashError,
};

export default connect(mapState, mapDispatch)(Home);
