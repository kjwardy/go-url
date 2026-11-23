import axios from 'axios';

type FrontendComponent =
  | 'home'
  | 'search'
  | 'history'
  | 'metrics'
  | 'most_wanted'
  | 'edit_url'
  | 'delete_url';

type FrontendOperation = 'url.create' | 'url.update' | 'url.delete';
type OperationOutcome = 'success' | 'failure';

const report = (event: Record<string, string | number>) => {
  void axios
    .post('/api/frontend-events', { version: 1, ...event })
    .catch(() => undefined);
};

export const reportFrontendError = (component: FrontendComponent) => {
  report({ type: 'error', component, code: 'api_request_failed' });
};

export const reportFrontendOperation = (
  operation: FrontendOperation,
  outcome: OperationOutcome,
) => {
  report({ type: 'operation', operation, outcome });
};
