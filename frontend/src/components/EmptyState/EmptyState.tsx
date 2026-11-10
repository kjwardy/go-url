import React from 'react';
import Button from '@material-ui/core/Button';
import CircularProgress from '@material-ui/core/CircularProgress';
import Typography from '@material-ui/core/Typography';

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
  <div
    style={{ padding: 32, textAlign: 'center' }}
    role={loading ? 'status' : 'alert'}
  >
    {loading && <CircularProgress size={24} aria-label="Loading" />}
    <Typography variant="subtitle1">{title}</Typography>
    <Typography variant="body2" color="textSecondary">
      {description}
    </Typography>
    {onRetry && (
      <Button color="primary" onClick={onRetry}>
        Try again
      </Button>
    )}
  </div>
);

export default EmptyState;
