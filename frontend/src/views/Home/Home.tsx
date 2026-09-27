import React, { useEffect, useState } from 'react';
import axios from 'axios';
import qs from 'qs';
import { connect } from 'react-redux';
import { useLocation } from 'react-router-dom';
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
    created?: any[];
    updated?: any[];
    deleted?: string[];
  };
}

const Home: React.FC<HomeProps> = ({ search, displayFlashError }) => {
  const [popular, setPopular] = useState<any[]>();
  const [popularStatus, setPopularStatus] = useState<
    'loading' | 'loaded' | 'error'
  >('loading');
  const [popularRetryCount, setPopularRetryCount] = useState(0);
  const location = useLocation();

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
  const deleted = search.deleted || [];
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
  const popularResults = sortByViews(
    applyUpdates(addCreated(popular)).filter(
      (result) => !deleted.includes(result.key),
    ),
  );

  return (
    <div className="mx-auto grid max-w-[1160px] grid-cols-1 gap-6 px-4 py-6 sm:px-6 sm:py-10 lg:grid-cols-[minmax(0,820px)_300px]">
      <main className="min-w-0">
        <Tabs defaultValue="popular" aria-label="URL data views">
          <TabsList>
            <TabsTrigger value="popular">Most Popular</TabsTrigger>
            <TabsTrigger value="wanted">Most Wanted</TabsTrigger>
            <TabsTrigger value="history">History</TabsTrigger>
          </TabsList>
          <TabsContent value="popular">
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
