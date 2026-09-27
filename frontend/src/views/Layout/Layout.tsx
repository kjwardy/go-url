import React, { useState, useCallback } from 'react';
import { connect } from 'react-redux';
import { useLocation } from 'react-router-dom';
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
  const hideAdd = useCallback(() => setAddOpen(false), []);
  const location = useLocation();
  // Pre-populate field if not found
  const urlQuery =
    location.search.includes('message=') &&
    decodeURIComponent(location.pathname.slice(1));

  return (
    <div className="min-h-screen bg-background text-foreground">
      {flash.message && (
        <Alert variant={flash.variant} message={flash.message} />
      )}
      {addOpen && (
        <EditModal onClose={hideAdd} urlKey={urlQuery || undefined} />
      )}
      <Header mode={mode} onToggleMode={onToggleMode} />
      {children}
      <Tooltip>
        <TooltipTrigger
          render={
            <Button
              data-e2e="add-button"
              variant="secondary"
              size="icon-lg"
              aria-label="Add New URL"
              className="fixed bottom-6 right-6 size-16 rounded-full shadow-md"
              onClick={() => setAddOpen(true)}
            />
          }
        >
          <Plus className="h-9 w-9" />
        </TooltipTrigger>
        <TooltipContent>Add New URL</TooltipContent>
      </Tooltip>
    </div>
  );
};

const mapState = ({ flash }) => ({ flash });
export default connect(mapState)(Layout);
