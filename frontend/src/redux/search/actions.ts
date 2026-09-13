import { SEARCH, URL_CREATED } from './constants';

// eslint-disable-next-line import/prefer-default-export
export const searchResults = (data: any) => ({
  type: SEARCH,
  data,
});

export const urlCreated = (data: any) => ({
  type: URL_CREATED,
  data,
});
