import React from 'react';
import { Loader2, RotateCcw } from 'lucide-react';
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
  EmptyContent,
} from '../ui/empty';
import { Button } from '../ui/button';

interface EmptyStateProps {
  title: string;
  description: string;
  loading?: boolean;
  onRetry?: () => void;
}

const EmptyState: React.FC<EmptyStateProps> = ({
  title,
  description,
  loading = false,
  onRetry,
}) => (
  <Empty className="min-h-36 py-8" role="status" aria-live="polite">
    <EmptyHeader>
      {loading && (
        <EmptyMedia variant="icon">
          <Loader2 className="animate-spin" aria-hidden="true" />
        </EmptyMedia>
      )}
      <EmptyTitle>{title}</EmptyTitle>
      <EmptyDescription>{description}</EmptyDescription>
    </EmptyHeader>
    {onRetry && (
      <EmptyContent>
        <Button type="button" variant="outline" size="sm" onClick={onRetry}>
          <RotateCcw aria-hidden="true" data-icon="inline-start" />
          Retry
        </Button>
      </EmptyContent>
    )}
  </Empty>
);

export default EmptyState;
