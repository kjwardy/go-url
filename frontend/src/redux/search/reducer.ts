import { SEARCH, URL_CREATED, URL_UPDATED } from './constants';

const initialState = {
  created: [] as any[],
  updated: [] as any[],
  results: undefined as any,
};

interface IAction {
  type: string;
  data: any;
}

const reducer = (state = { ...initialState }, action: IAction) => {
  switch (action.type) {
    case SEARCH:
      return {
        ...state,
        results: action.data,
      };
    // Store new URLs so results can update without another API request
    case URL_CREATED:
      return {
        ...state,
        created: [action.data, ...(state.created || [])],
      };
    // Store the latest value for each updated URL
    case URL_UPDATED:
      return {
        ...state,
        updated: [
          action.data,
          ...state.updated.filter((url) => url.key !== action.data.key),
        ],
      };
    default:
      return state;
  }
};

export default reducer;
