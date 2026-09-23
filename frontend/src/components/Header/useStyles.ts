import { Theme } from '@material-ui/core/styles';
import { makeStyles } from '@material-ui/styles';

const useStyles = makeStyles((theme: Theme) => ({
  grow: {
    flexGrow: 1,
  },
  name: {
    marginLeft: theme.spacing(2),
    padding: theme.spacing(0.75, 1.25),
    borderRadius: 20,
    backgroundColor: 'rgba(255, 255, 255, 0.1)',
    fontWeight: 500,
  },
  logout: {
    display: 'flex',
    marginLeft: theme.spacing(0.5),
  },
  link: {
    display: 'flex',
    alignItems: 'center',
    padding: theme.spacing(1),
    color: theme.palette.common.white,
    textDecoration: 'none',
    fontWeight: 600,
    borderRadius: theme.shape.borderRadius,
    transition: theme.transitions.create('background-color'),
    '&:hover': {
      backgroundColor: 'rgba(255, 255, 255, 0.08)',
    },
  },
  linkSecondary: {
    marginLeft: theme.spacing(2),
  },
  logo: {
    height: 34,
    marginRight: theme.spacing(1),
    filter: 'drop-shadow(0 2px 4px rgba(0, 0, 0, 0.18))',
  },
}));

export default useStyles;
