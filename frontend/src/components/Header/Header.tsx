import React, { useState, useEffect } from 'react';
import Cookies from 'js-cookie';
import { LogOut, Moon, Sun } from 'lucide-react';
import Search from '../Search';
import { Button } from '../ui/button';
import { Tooltip, TooltipContent, TooltipTrigger } from '../ui/tooltip';

interface HeaderProps {
  onSearch: (query: string) => void;
  mode: 'light' | 'dark';
  onToggleMode: () => void;
}

const Header: React.FC<HeaderProps> = ({ onSearch, mode, onToggleMode }) => {
  const [name, setName] = useState('');
  useEffect(() => {
    const name = Cookies.get('user');
    if (name) {
      setName(name);
    }
  }, []);
  return (
    <header className="bg-[#33469b] text-white shadow-sm dark:bg-card">
      <div className="mx-auto flex min-h-16 max-w-7xl items-center gap-2 px-4 sm:px-6">
        <a
          className="flex shrink-0 items-center rounded-md p-2 font-semibold text-white no-underline transition-colors hover:bg-white/10"
          href="/go"
        >
          <img
            src={process.env.PUBLIC_URL + '/logo.svg'}
            alt="Go URL Logo"
            className="mr-2 h-[34px] drop-shadow"
          />
          <span className="text-lg">Go</span>
        </a>
        <a
          className="ml-2 rounded-md p-2 font-medium text-white no-underline transition-colors hover:bg-white/10"
          href="/help"
        >
          Help
        </a>
        <div className="flex-1" />
        <Search onSearch={onSearch} />
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant="ghost"
              size="icon"
              className="shrink-0 text-white hover:bg-white/10 hover:text-white"
              aria-label="Toggle dark mode"
              data-e2e="theme-toggle"
              onClick={onToggleMode}
            >
              {mode === 'light' ? (
                <Moon className="h-5 w-5" />
              ) : (
                <Sun className="h-5 w-5" />
              )}
            </Button>
          </TooltipTrigger>
          <TooltipContent>
            Switch to {mode === 'light' ? 'dark' : 'light'} mode
          </TooltipContent>
        </Tooltip>

        {name && (
          <>
            <span className="ml-2 rounded-full bg-white/10 px-3 py-1.5 font-medium">
              {name}
            </span>
            <form method="post" action="/logout" className="ml-1 flex">
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="text-white hover:bg-white/10 hover:text-white"
                    aria-label="Log out"
                    type="submit"
                  >
                    <LogOut className="h-5 w-5" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>Log out</TooltipContent>
              </Tooltip>
            </form>
          </>
        )}
      </div>
    </header>
  );
};

export default Header;
