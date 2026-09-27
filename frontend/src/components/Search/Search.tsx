import React, { useState } from 'react';
import { Search as SearchIcon } from 'lucide-react';
import { Input } from '../ui/input';

interface SearchProps {
  onSearch: (query: string) => void;
}

const Search: React.FC<SearchProps> = ({ onSearch }) => {
  const [query, setQuery] = useState('');
  return (
    <form
      role="search"
      className="relative w-full rounded-full border border-white/20 bg-white/10 transition-colors hover:border-white/40 hover:bg-white/15 focus-within:border-white/40 focus-within:bg-white/15 sm:ml-2 sm:w-auto"
      onSubmit={(event) => {
        event.preventDefault();
        onSearch(query);
      }}
    >
      <SearchIcon className="pointer-events-none absolute left-4 top-1/2 h-4 w-4 -translate-y-1/2 text-white/80" />
      <Input
        aria-label="Search URLs"
        placeholder="Search…"
        onChange={(e) => setQuery(e.target.value)}
        value={query}
        className="h-10 w-full rounded-full border-0 bg-transparent pl-11 text-white placeholder:text-white/70 focus-visible:ring-0 focus-visible:ring-offset-0 sm:w-[280px] sm:focus-visible:w-[360px]"
      />
    </form>
  );
};

export default Search;
