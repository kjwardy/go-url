import { green } from '@material-ui/core/colors';
import { Theme } from '@material-ui/core/styles';
import { makeStyles } from '@material-ui/styles';

const useStyles = makeStyles((theme: Theme) => ({
  paper: {
    padding: 15,
    overflowX: 'auto',
  },
  header: {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
  },
  successful: {
    color: green[600],
  },
  unsuccessful: {
    color: theme.palette.error.main,
  },
}));

export default useStyles;
