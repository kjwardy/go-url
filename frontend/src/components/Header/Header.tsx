import React, { useState, useEffect } from 'react';
import Cookies from 'js-cookie';
import { LogOut, Moon, Sun } from 'lucide-react';
import Search from '../Search';
import { Button } from '../ui/button';
import { Tooltip, TooltipContent, TooltipTrigger } from '../ui/tooltip';

interface HeaderProps {
  mode: 'light' | 'dark';
  onToggleMode: () => void;
}

const Header: React.FC<HeaderProps> = ({ mode, onToggleMode }) => {
  const [name, setName] = useState('');
  useEffect(() => {
    const name = Cookies.get('user');
    if (name) {
      setName(name);
    }
  }, []);
  return (
    <header className="border-b bg-primary text-primary-foreground dark:bg-card dark:text-card-foreground">
      <div className="mx-auto flex min-h-16 max-w-7xl items-center gap-2 px-4 sm:px-6">
        <Button
          variant="ghost"
          render={
            <a href="/go" className="shrink-0 font-semibold">
              <img
                src={process.env.PUBLIC_URL + '/logo.svg'}
                alt=""
                className="mr-2 h-8"
              />
              Go
            </a>
          }
        />
        <Button variant="ghost" render={<a href="/help">Help</a>} />
        <div className="flex-1" />
        <Search />
        <Tooltip>
          <TooltipTrigger
            render={
              <Button
                variant="ghost"
                size="icon"
                className="shrink-0"
                aria-label="Toggle dark mode"
                data-e2e="theme-toggle"
                onClick={onToggleMode}
              />
            }
          >
            {mode === 'light' ? (
              <Moon className="h-5 w-5" />
            ) : (
              <Sun className="h-5 w-5" />
            )}
          </TooltipTrigger>
          <TooltipContent>
            Switch to {mode === 'light' ? 'dark' : 'light'} mode
          </TooltipContent>
        </Tooltip>

        {name && (
          <>
            <span className="ml-2 text-sm font-medium">{name}</span>
            <form method="post" action="/logout" className="ml-1 flex">
              <Tooltip>
                <TooltipTrigger
                  render={
                    <Button
                      variant="ghost"
                      size="icon"
                      aria-label="Log out"
                      type="submit"
                    />
                  }
                >
                  <LogOut className="h-5 w-5" />
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
