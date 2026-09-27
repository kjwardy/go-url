import axios from 'axios';
import React, { useCallback, useEffect, useRef, useState } from 'react';
import { connect } from 'react-redux';
import { Loader2, Search as SearchIcon } from 'lucide-react';
import DeleteModal from '../DeleteModal';
import EditModal from '../EditModal';
import { Input } from '../ui/input';
import { Popover, PopoverContent } from '../ui/popover';

interface SearchResult {
  key: string;
  url: string;
  alias?: string[];
  views: number;
}

const SEARCH_DEBOUNCE_MS = 300;

interface SearchProps {
  deleted: string[];
}

const Search: React.FC<SearchProps> = ({ deleted }) => {
  const [query, setQuery] = useState('');
  const [debouncedQuery, setDebouncedQuery] = useState('');
  const [results, setResults] = useState<SearchResult[]>([]);
  const [selectedResult, setSelectedResult] = useState<SearchResult | null>(
    null,
  );
  const [deleteSelected, setDeleteSelected] = useState<SearchResult | null>(
    null,
  );
  const [searchStatus, setSearchStatus] = useState<
    'idle' | 'loading' | 'success' | 'error'
  >('idle');
  const [isFocused, setIsFocused] = useState(false);
  const formRef = useRef<HTMLFormElement>(null);
  const trimmedQuery = query.trim();
  const isDebouncing = trimmedQuery !== debouncedQuery;
  const isOpen =
    isFocused && trimmedQuery.length > 0 && !selectedResult && !deleteSelected;
  const visibleResults = results.filter(
    (result) => !deleted.includes(result.key),
  );
  const closeEdit = useCallback(() => {
    setSelectedResult(null);
    setIsFocused(false);
  }, []);
  const closeDelete = useCallback(() => {
    setDeleteSelected(null);
    setIsFocused(false);
  }, []);
  const startDelete = useCallback(() => {
    if (!selectedResult) return;
    setDeleteSelected(selectedResult);
    setIsFocused(false);
  }, [selectedResult]);

  useEffect(() => {
    const timeout = window.setTimeout(
      () => setDebouncedQuery(trimmedQuery),
      SEARCH_DEBOUNCE_MS,
    );
    return () => window.clearTimeout(timeout);
  }, [trimmedQuery]);

  useEffect(() => {
    if (!debouncedQuery) {
      setResults([]);
      setSearchStatus('idle');
      return undefined;
    }

    let isCurrent = true;
    setSearchStatus('loading');
    axios
      .get<SearchResult[]>('/api/search', { params: { q: debouncedQuery } })
      .then(({ data }) => {
        if (!isCurrent) return;
        setResults(data);
        setSearchStatus('success');
      })
      .catch(() => {
        if (!isCurrent) return;
        setSearchStatus('error');
      });

    return () => {
      isCurrent = false;
    };
  }, [debouncedQuery]);

  return (
    <>
      <form
        ref={formRef}
        role="search"
        className="relative w-full sm:ml-2 sm:w-auto"
        onSubmit={(event) => {
          event.preventDefault();
          if (searchStatus === 'success' && visibleResults[0]) {
            setSelectedResult(visibleResults[0]);
            setIsFocused(false);
          }
        }}
      >
        <Popover open={isOpen} onOpenChange={(open) => setIsFocused(open)}>
          <SearchIcon className="pointer-events-none absolute left-3 top-1/2 z-10 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            aria-label="Search URLs"
            aria-autocomplete="list"
            aria-controls="search-suggestions"
            aria-expanded={isOpen}
            autoComplete="off"
            placeholder="Search…"
            onChange={(event) => setQuery(event.target.value)}
            onFocus={() => setIsFocused(true)}
            onBlur={() => {
              if (!trimmedQuery) setIsFocused(false);
            }}
            value={query}
            className="pl-9 sm:w-72"
          />
          <PopoverContent
            anchor={formRef}
            initialFocus={false}
            finalFocus={false}
            className="max-h-80 overflow-y-auto"
          >
            {isDebouncing || searchStatus === 'loading' ? (
              <div
                className="flex items-center justify-center gap-2 px-4 py-6 text-sm text-muted-foreground"
                role="status"
              >
                <Loader2 className="h-4 w-4 animate-spin" aria-hidden="true" />
                Searching…
              </div>
            ) : searchStatus === 'error' ? (
              <div className="px-4 py-6 text-center" role="status">
                <p className="text-sm font-medium">Search unavailable</p>
                <p className="mt-1 text-sm text-muted-foreground">
                  Try again in a moment.
                </p>
              </div>
            ) : visibleResults.length === 0 ? (
              <div className="px-4 py-6 text-center" role="status">
                <p className="text-sm font-medium">No results found</p>
                <p className="mt-1 text-sm text-muted-foreground">
                  Try a different search term.
                </p>
              </div>
            ) : (
              <ul
                id="search-suggestions"
                aria-label="Search results"
                role="listbox"
              >
                {visibleResults.map((result) => (
                  <li key={result.key} role="option" aria-selected={false}>
                    <button
                      type="button"
                      className="block w-full border-b px-4 py-3 text-left last:border-b-0 hover:bg-accent focus-visible:bg-accent focus-visible:outline-none"
                      onClick={() => {
                        setSelectedResult(result);
                        setIsFocused(false);
                      }}
                    >
                      <span className="block truncate text-sm font-medium">
                        {result.key}
                      </span>
                      <span className="mt-1 block truncate text-xs text-muted-foreground">
                        {result.alias?.length
                          ? result.alias.join(', ')
                          : result.url}
                      </span>
                    </button>
                  </li>
                ))}
              </ul>
            )}
          </PopoverContent>
        </Popover>
      </form>
      {selectedResult && !deleteSelected && (
        <EditModal
          edit
          urlKey={selectedResult.key}
          url={selectedResult.url || selectedResult.alias?.join(',') || ''}
          onClose={closeEdit}
          onDelete={startDelete}
        />
      )}
      {deleteSelected && (
        <DeleteModal
          urlKey={deleteSelected.key}
          onClose={closeDelete}
          onDeleted={() => {
            setResults((current) =>
              current.filter((result) => result.key !== deleteSelected.key),
            );
            setSelectedResult(null);
          }}
        />
      )}
    </>
  );
};

const mapState = ({ search }) => ({ deleted: search.deleted || [] });

export default connect(mapState)(Search);
