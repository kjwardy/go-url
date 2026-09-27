import { SEARCH, URL_CREATED, URL_UPDATED, URL_DELETED } from './constants';

// eslint-disable-next-line import/prefer-default-export
export const searchResults = (data: any) => ({
  type: SEARCH,
  data,
});

export const urlCreated = (data: any) => ({
  type: URL_CREATED,
  data,
});

export const urlUpdated = (data: any) => ({
  type: URL_UPDATED,
  data,
});

export const urlDeleted = (key: string) => ({
  type: URL_DELETED,
  data: key,
});
