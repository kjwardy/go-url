import React, { useState, useCallback } from 'react';
import { ArrowUpRight, Pencil, Trash2 } from 'lucide-react';

import DeleteModal from '../DeleteModal';
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

interface IResult {
  key: string;
  url: string;
  alias: string[];
  views: number;
}

interface ResultsProps {
  data: IResult[];
  title: string;
  emptyState?: {
    title: string;
    description: string;
    loading?: boolean;
    onRetry?: () => void;
  };
}

const Results: React.FC<ResultsProps> = ({ data, title, emptyState }) => {
  const [selected, setSelected] = useState<IResult | null>(null);
  const [deleteSelected, setDeleteSelected] = useState<IResult | null>(null);
  // Track deleted keys locally so rows disappear without reloading the page
  const [deletedKeys, setDeletedKeys] = useState<string[]>([]);
  const clearSelected = useCallback(() => setSelected(null), []);
  const clearDeleteSelected = useCallback(() => setDeleteSelected(null), []);
  const results = data.filter((result) => !deletedKeys.includes(result.key));

  const getFormattedUrl = (url: string) => {
    const regex = /({{\$\d+}})/g;
    const parts = url.split(regex);
    return parts.map((part, i) =>
      part.match(regex) ? (
        <span key={i} className="font-bold text-foreground">
          {part}
        </span>
      ) : (
        part
      ),
    );
  };
  return (
    <div>
      {selected && (
        <EditModal
          edit
          urlKey={selected.key}
          url={selected.url || selected.alias.join(',')}
          onClose={clearSelected}
        />
      )}
      {deleteSelected && (
        <DeleteModal
          urlKey={deleteSelected.key}
          onClose={clearDeleteSelected}
          onDeleted={() =>
            setDeletedKeys((keys) => [...keys, deleteSelected.key])
          }
        />
      )}

      <Card data-e2e={title}>
        <CardHeader>
          <CardTitle>{title}</CardTitle>
        </CardHeader>
        <CardContent>
          {!results.length ? (
            <EmptyState
              title={emptyState?.title || 'No results found'}
              description={
                emptyState?.description || 'Help others by adding it.'
              }
              loading={emptyState?.loading}
              onRetry={emptyState?.onRetry}
            />
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Key</TableHead>
                  <TableHead>Url</TableHead>
                  <TableHead className="text-right">Views</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {results.map((r) => (
                  <TableRow key={r.key}>
                    <TableCell>{r.key}</TableCell>
                    <TableCell className="max-w-[300px] break-words max-sm:max-w-[100px]">
                      {r.alias && r.alias.length ? (
                        r.alias.map((alias) => (
                          <a
                            key={alias}
                            className="font-medium text-primary hover:underline"
                            href={`/${encodeURIComponent(alias)}`}
                          >
                            {alias}
                            <ArrowUpRight
                              aria-hidden="true"
                              className="ml-0.5 inline h-3 w-3"
                            />
                          </a>
                        ))
                      ) : (
                        <a
                          className="font-medium text-primary hover:underline"
                          href={`/${encodeURIComponent(r.key)}`}
                        >
                          {getFormattedUrl(r.url)}
                          <ArrowUpRight
                            aria-hidden="true"
                            className="ml-0.5 inline h-3 w-3"
                          />
                        </a>
                      )}
                    </TableCell>
                    <TableCell className="text-right">{r.views}</TableCell>
                    <TableCell className="whitespace-nowrap text-right">
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-9 w-9 text-primary"
                        onClick={() => setSelected(r)}
                        aria-label={`Edit ${r.key}`}
                        data-e2e="edit"
                      >
                        <Pencil className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-9 w-9 text-destructive"
                        onClick={() => setDeleteSelected(r)}
                        aria-label={`Delete ${r.key}`}
                        data-e2e="delete"
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  );
};

export default Results;
