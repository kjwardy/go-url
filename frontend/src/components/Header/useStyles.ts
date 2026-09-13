import { Theme } from '@material-ui/core/styles';
import { makeStyles } from '@material-ui/styles';

const useStyles = makeStyles((theme: Theme) => ({
  grow: {
    flexGrow: 1,
  },
  name: {
    marginLeft: 20,
    fontWeight: 500,
  },
  link: {
    textDecoration: 'none',
    padding: 10,
    color: theme.palette.primary.contrastText,
    marginLeft: 30,
    fontWeight: 600,
  },
}));

export default useStyles;
