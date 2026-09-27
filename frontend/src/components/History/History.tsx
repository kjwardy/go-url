import React, { useEffect, useState } from 'react';
import axios from 'axios';
import { Check, X } from 'lucide-react';
import { Card } from '../ui/card';
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

  useEffect(() => {
    axios
      .get<HistoryEntry[]>('/api/history', { params: { limit } })
      .then(({ data }) => setHistory(data))
      .catch((err) =>
        displayFlashError(err.response.data.message || err.response.data),
      );
  }, [limit, displayFlashError]);

  return (
    <Card className="overflow-x-auto p-5 shadow-[0_10px_35px_rgba(20,29,60,0.08)]">
      <div className="mb-4 flex items-center justify-between gap-4">
        <h3 className="text-lg font-semibold">Query History</h3>
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
      </div>
      {history && history.length === 0 ? (
        <p className="text-sm text-muted-foreground">
          No query history found - go nuts!
        </p>
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
                      <TooltipTrigger asChild>
                        <Check
                          className="inline h-4 w-4 text-emerald-600"
                          aria-label="Resolved successfully"
                        />
                      </TooltipTrigger>
                      <TooltipContent>Resolved successfully</TooltipContent>
                    </Tooltip>
                  ) : (
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <X
                          className="inline h-4 w-4 text-destructive"
                          aria-label="Did not resolve"
                        />
                      </TooltipTrigger>
                      <TooltipContent>Did not resolve</TooltipContent>
                    </Tooltip>
                  )}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}
    </Card>
  );
};

export default History;
