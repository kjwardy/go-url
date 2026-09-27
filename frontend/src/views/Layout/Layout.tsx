import React, { useState, useEffect, useCallback } from 'react';
import { connect } from 'react-redux';
import { useHistory, useLocation } from 'react-router-dom';
import { Plus } from 'lucide-react';
import EditModal from '../../components/EditModal';
import Header from '../../components/Header';
import Alert from '../../components/Alert';
import type { Variant } from '../../components/Alert/Alert';
import { Button } from '../../components/ui/button';
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '../../components/ui/tooltip';

interface LayoutProps {
  children?: React.ReactNode;
  mode: 'light' | 'dark';
  onToggleMode: () => void;
  flash: {
    message: string;
    variant: Variant;
  };
}

const Layout: React.FC<LayoutProps> = ({
  children,
  flash,
  mode,
  onToggleMode,
}) => {
  const [addOpen, setAddOpen] = useState(false);
  const [query, onSearch] = useState<string>();
  const hideAdd = useCallback(() => setAddOpen(false), []);
  const history = useHistory();
  const location = useLocation();
  // Pre-populate field if not found
  const urlQuery =
    location.search.includes('message=') &&
    decodeURIComponent(location.pathname.slice(1));

  useEffect(() => {
    if (query === undefined) return;
    history.push(`/${encodeURIComponent(query)}`);
  }, [query, history]);

  return (
    <div className="min-h-screen bg-background bg-[radial-gradient(circle_at_85%_0%,rgba(64,84,178,0.08),transparent_28%)] text-foreground dark:bg-[radial-gradient(circle_at_85%_0%,rgba(142,162,255,0.08),transparent_28%)]">
      {flash.message && (
        <Alert variant={flash.variant} message={flash.message} />
      )}
      {addOpen && (
        <EditModal onClose={hideAdd} urlKey={urlQuery || undefined} />
      )}
      <Header onSearch={onSearch} mode={mode} onToggleMode={onToggleMode} />
      {children}
      <Tooltip>
        <TooltipTrigger asChild>
          <Button
            data-e2e="add-button"
            variant="secondary"
            size="icon"
            aria-label="Add New URL"
            className="fixed bottom-6 right-6 h-14 w-14 rounded-full shadow-lg"
            onClick={() => setAddOpen(true)}
          >
            <Plus className="h-6 w-6" />
          </Button>
        </TooltipTrigger>
        <TooltipContent>Add New URL</TooltipContent>
      </Tooltip>
    </div>
  );
};

const mapState = ({ flash }) => ({ flash });
export default connect(mapState)(Layout);
