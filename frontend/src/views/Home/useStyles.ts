import { Theme } from '@material-ui/core/styles';
import { makeStyles } from '@material-ui/styles';

const useStyles = makeStyles((theme: Theme) => ({
  dashboard: {
    display: 'grid',
    gridTemplateColumns: 'minmax(0, 800px) 280px',
    gap: theme.spacing(2.5),
    maxWidth: 1100,
    margin: `${theme.spacing(2.5)}px auto`,
    padding: `0 ${theme.spacing(2)}px`,
    [theme.breakpoints.down('sm')]: {
      gridTemplateColumns: 'minmax(0, 1fr)',
    },
  },
  main: {
    minWidth: 0,
  },
  container: {
    marginBottom: theme.spacing(2.5),
  },
  tabs: {
    marginBottom: theme.spacing(2.5),
  },
}));

export default useStyles;
