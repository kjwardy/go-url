import React, { useEffect, useState } from 'react';
import axios from 'axios';
import qs from 'qs';
import { connect } from 'react-redux';
import { useRouteMatch, useLocation } from 'react-router-dom';
import History from '../../components/History';
import Metrics from '../../components/Metrics';
import MostWanted from '../../components/MostWanted';
import Results from '../../components/Results';
import { displayFlashError } from '../../redux/flash/actions';
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from '../../components/ui/tabs';

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
    <div className="mx-auto grid max-w-[1160px] grid-cols-1 gap-6 px-4 py-6 sm:px-6 sm:py-10 lg:grid-cols-[minmax(0,820px)_300px]">
      <main className="min-w-0">
        {searchResults && (
          <div className="mb-6">
            <Results
              data={applyUpdates(
                addCreated(searchResults, createdSearchResults),
              )}
              title="Search Results"
            />
          </div>
        )}
        <Tabs defaultValue="popular" aria-label="URL data views">
          <TabsList>
            <TabsTrigger value="popular">Most Popular</TabsTrigger>
            <TabsTrigger value="wanted">Most Wanted</TabsTrigger>
            <TabsTrigger value="history">History</TabsTrigger>
          </TabsList>
          <TabsContent value="popular">
            {Boolean(popular || created.length > 0) && (
              <Results
                data={sortByViews(applyUpdates(addCreated(popular)))}
                title="Most Popular"
              />
            )}
          </TabsContent>
          <TabsContent value="wanted">
            <MostWanted displayFlashError={displayFlashError} />
          </TabsContent>
          <TabsContent value="history">
            <History displayFlashError={displayFlashError} />
          </TabsContent>
        </Tabs>
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
