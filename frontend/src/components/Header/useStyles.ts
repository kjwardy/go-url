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
    fontWeight: 600,
    display: 'flex',
    alignItems: 'center',
  },
  linkSecondary: {
    marginLeft: 30,
  },
  logo: {
    height: 32,
    marginRight: 8,
  },
}));

export default useStyles;
