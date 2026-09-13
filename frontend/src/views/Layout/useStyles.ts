import { Theme } from '@material-ui/core/styles';
import { makeStyles } from '@material-ui/styles';

const useStyles = makeStyles((theme: Theme) => ({
  root: {
    minHeight: '100vh',
    color: theme.palette.text.primary,
    backgroundColor: theme.palette.background.default,
  },
  button: {
    position: 'fixed',
    right: 23,
    bottom: 23,
  },
}));

export default useStyles;
