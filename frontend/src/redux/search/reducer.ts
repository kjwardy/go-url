import { SEARCH, URL_CREATED } from './constants';

const initialState = {
  created: [] as any[],
  results: undefined as any,
};

interface IAction {
  type: string;
  data?: any;
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
    default:
      return state;
  }
};

export default reducer;
