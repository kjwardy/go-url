import { Theme } from '@material-ui/core/styles';
import { makeStyles } from '@material-ui/styles';

const useStyles = makeStyles((theme: Theme) => ({
  dashboard: {
    display: 'grid',
    gridTemplateColumns: 'minmax(0, 820px) 300px',
    gap: theme.spacing(3),
    maxWidth: 1160,
    margin: `${theme.spacing(4)}px auto`,
    padding: `0 ${theme.spacing(3)}px`,
    [theme.breakpoints.down('sm')]: {
      gridTemplateColumns: 'minmax(0, 1fr)',
      marginTop: theme.spacing(2.5),
      padding: `0 ${theme.spacing(2)}px`,
    },
  },
  main: {
    minWidth: 0,
  },
  container: {
    marginBottom: theme.spacing(3),
  },
  tabs: {
    marginBottom: theme.spacing(3),
    padding: theme.spacing(0.5),
    border: `1px solid ${theme.palette.divider}`,
    boxShadow: 'none',
    '& .MuiTabs-indicator': {
      height: 3,
      borderRadius: 3,
    },
  },
}));

export default useStyles;
