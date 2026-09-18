import { Theme } from '@material-ui/core/styles';
import { makeStyles } from '@material-ui/styles';

const useStyles = makeStyles((theme: Theme) => ({
  paper: {
    padding: theme.spacing(2.5),
    overflowX: 'auto',
    border: `1px solid ${theme.palette.divider}`,
    boxShadow: '0 10px 35px rgba(20, 29, 60, 0.08)',
  },
  header: {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
    gap: theme.spacing(2),
    marginBottom: theme.spacing(2),
    '& h3': {
      margin: 0,
      fontSize: '1.15rem',
      fontWeight: 600,
    },
  },
  successful: {
    color: theme.palette.success.main,
  },
  unsuccessful: {
    color: theme.palette.error.main,
  },
}));

export default useStyles;
