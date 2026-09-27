import React, { useCallback, useEffect, useState } from 'react';
import axios from 'axios';
import { Pencil } from 'lucide-react';
import EditModal from '../EditModal';
import EmptyState from '../EmptyState/EmptyState';
import { Button } from '../ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '../ui/card';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '../ui/table';

interface MostWantedEntry {
  query: string;
  views: number;
}

interface MostWantedProps {
  displayFlashError: (message: string) => void;
}

const MostWanted: React.FC<MostWantedProps> = ({ displayFlashError }) => {
  const [mostWanted, setMostWanted] = useState<MostWantedEntry[]>();
  const [selected, setSelected] = useState<MostWantedEntry>();
  const [loadStatus, setLoadStatus] = useState<'loading' | 'loaded' | 'error'>(
    'loading',
  );
  const fetchMostWanted = useCallback(() => {
    setLoadStatus('loading');
    axios
      .get<MostWantedEntry[]>('/api/most-wanted')
      .then(({ data }) => {
        setMostWanted(data);
        setLoadStatus('loaded');
      })
      .catch((err) => {
        setLoadStatus('error');
        displayFlashError(err.response.data.message || err.response.data);
      });
  }, [displayFlashError]);

  useEffect(() => {
    fetchMostWanted();
  }, [fetchMostWanted]);

  return (
    <Card>
      {selected && (
        <EditModal
          urlKey={selected.query}
          onClose={() => setSelected(undefined)}
          onCreated={() => {
            setMostWanted((entries) =>
              entries?.filter(({ query }) => query !== selected.query),
            );
          }}
        />
      )}
      <CardHeader>
        <CardTitle>Most Wanted</CardTitle>
      </CardHeader>
      <CardContent>
        {loadStatus === 'loading' ? (
          <EmptyState
            title="Loading unresolved queries"
            description="Queries without matching URLs will appear here."
            loading
          />
        ) : loadStatus === 'error' ? (
          <EmptyState
            title="Could not load unresolved queries"
            description="Try again later. The request failed before the list could be loaded."
            onRetry={fetchMostWanted}
          />
        ) : mostWanted?.length === 0 ? (
          <EmptyState
            title="No unresolved queries"
            description="Queries without matching URLs will appear here."
          />
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Query</TableHead>
                <TableHead className="text-right">Views</TableHead>
                <TableHead className="text-center">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {mostWanted?.map((entry) => (
                <TableRow key={entry.query}>
                  <TableCell>{entry.query}</TableCell>
                  <TableCell className="text-right">{entry.views}</TableCell>
                  <TableCell className="text-center">
                    <Button
                      variant="ghost"
                      size="icon"
                      aria-label={`Add URL for ${entry.query}`}
                      onClick={() => setSelected(entry)}
                    >
                      <Pencil className="h-4 w-4" />
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </CardContent>
    </Card>
  );
};

export default MostWanted;
