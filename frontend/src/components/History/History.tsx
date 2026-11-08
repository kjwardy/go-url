import React, { useEffect, useState } from 'react';
import axios from 'axios';
import { Check, X } from 'lucide-react';
import EmptyState from '../EmptyState/EmptyState';
import { Card, CardContent, CardHeader, CardTitle } from '../ui/card';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '../ui/select';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '../ui/table';
import { Tooltip, TooltipContent, TooltipTrigger } from '../ui/tooltip';

type HistoryLimit = 25 | 50 | 100;

interface HistoryEntry {
  id: number;
  url_key: string;
  queried_at: string;
  successful: boolean;
}

interface HistoryProps {
  displayFlashError: (message: string) => void;
}

const History: React.FC<HistoryProps> = ({ displayFlashError }) => {
  const [limit, setLimit] = useState<HistoryLimit>(25);
  const [history, setHistory] = useState<HistoryEntry[]>();
  const [loadStatus, setLoadStatus] = useState<'loading' | 'loaded' | 'error'>(
    'loading',
  );
  const [retryCount, setRetryCount] = useState(0);

  useEffect(() => {
    let cancelled = false;
    setLoadStatus('loading');
    axios
      .get<HistoryEntry[]>('/api/history', { params: { limit } })
      .then(({ data }) => {
        if (cancelled) return;
        setHistory(data);
        setLoadStatus('loaded');
      })
      .catch((err) => {
        if (cancelled) return;
        setLoadStatus('error');
        displayFlashError(err.response.data.message || err.response.data);
      });
    return () => {
      cancelled = true;
    };
  }, [limit, displayFlashError, retryCount]);

  return (
    <Card>
      <CardHeader className="flex-row items-center justify-between space-y-0">
        <CardTitle>Query History</CardTitle>
        <Select
          value={String(limit)}
          onValueChange={(value) => setLimit(Number(value) as HistoryLimit)}
        >
          <SelectTrigger
            id="history-limit"
            aria-label="Records"
            className="w-28"
          >
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="25">25</SelectItem>
            <SelectItem value="50">50</SelectItem>
            <SelectItem value="100">100</SelectItem>
          </SelectContent>
        </Select>
      </CardHeader>
      <CardContent>
        {loadStatus === 'loading' ? (
          <EmptyState
            title="Loading query history"
            description="Your recent queries will appear here."
            loading
          />
        ) : loadStatus === 'error' ? (
          <EmptyState
            title="Could not load query history"
            description="Try again later. The request failed before the history could be loaded."
            onRetry={() => setRetryCount((count) => count + 1)}
          />
        ) : history?.length === 0 ? (
          <EmptyState
            title="No query history yet"
            description="Queries you make will appear here."
          />
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Timestamp</TableHead>
                <TableHead>Query</TableHead>
                <TableHead className="text-center">Successful</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {(history || []).map((entry) => (
                <TableRow key={entry.id}>
                  <TableCell>
                    {new Date(entry.queried_at).toLocaleString()}
                  </TableCell>
                  <TableCell>{entry.url_key}</TableCell>
                  <TableCell className="text-center">
                    {entry.successful ? (
                      <Tooltip>
                        <TooltipTrigger
                          render={
                            <Check
                              className="inline h-4 w-4 text-secondary-foreground"
                              aria-label="Resolved successfully"
                            />
                          }
                        />
                        <TooltipContent>Resolved successfully</TooltipContent>
                      </Tooltip>
                    ) : (
                      <Tooltip>
                        <TooltipTrigger
                          render={
                            <X
                              className="inline h-4 w-4 text-destructive"
                              aria-label="Did not resolve"
                            />
                          }
                        />
                        <TooltipContent>Did not resolve</TooltipContent>
                      </Tooltip>
                    )}
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

export default History;
