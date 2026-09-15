import { Theme } from '@material-ui/core/styles';
import { makeStyles } from '@material-ui/styles';

const useStyles = makeStyles((theme: Theme) => ({
  paper: {
    padding: theme.spacing(2),
  },
  editButton: {
    cursor: 'not-allowed',
  },
}));

export default useStyles;
