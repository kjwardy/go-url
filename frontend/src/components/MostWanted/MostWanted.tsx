import React, { useCallback, useEffect, useState } from 'react';
import axios from 'axios';
import { Pencil } from 'lucide-react';
import EditModal from '../EditModal';
import { Button } from '../ui/button';
import { Card } from '../ui/card';
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
  const fetchMostWanted = useCallback(() => {
    axios
      .get<MostWantedEntry[]>('/api/most-wanted')
      .then(({ data }) => setMostWanted(data))
      .catch((err) =>
        displayFlashError(err.response.data.message || err.response.data),
      );
  }, [displayFlashError]);

  useEffect(() => {
    fetchMostWanted();
  }, [fetchMostWanted]);

  return (
    <Card className="overflow-x-auto p-5 shadow-[0_10px_35px_rgba(20,29,60,0.08)]">
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
      <h3 className="mb-4 text-lg font-semibold">Most Wanted</h3>
      {mostWanted && mostWanted.length === 0 ? (
        <p className="text-sm text-muted-foreground">
          No unresolved queries found - go nuts!
        </p>
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
            {(mostWanted || []).map((entry) => (
              <TableRow key={entry.query}>
                <TableCell>{entry.query}</TableCell>
                <TableCell className="text-right">{entry.views}</TableCell>
                <TableCell className="text-center">
                  <Button
                    variant="ghost"
                    size="icon"
                    aria-label={`Add URL for ${entry.query}`}
                    className="h-9 w-9 text-primary"
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
    </Card>
  );
};

export default MostWanted;
