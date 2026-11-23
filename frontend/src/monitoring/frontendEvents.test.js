import axios from 'axios';
import {
  reportFrontendError,
  reportFrontendOperation,
} from './frontendEvents';

jest.mock('axios');

const post = axios.post;

describe('frontend event reporting', () => {
  beforeEach(() => {
    post.mockReset().mockResolvedValue({});
  });

  it('sends only an allowlisted error event', () => {
    reportFrontendError('search');

    expect(post).toHaveBeenCalledWith('/api/frontend-events', {
      version: 1,
      type: 'error',
      component: 'search',
      code: 'api_request_failed',
    });
  });

  it('sends allowlisted operation outcomes', () => {
    reportFrontendOperation('url.create', 'success');

    expect(post).toHaveBeenCalledWith('/api/frontend-events', {
      version: 1,
      type: 'operation',
      operation: 'url.create',
      outcome: 'success',
    });
  });

  it('does not propagate a reporting failure', async () => {
    post.mockRejectedValue(new Error('report failed'));

    expect(() => reportFrontendError('home')).not.toThrow();
    await Promise.resolve();
  });
});
