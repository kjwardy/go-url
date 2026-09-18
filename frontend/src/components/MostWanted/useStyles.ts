import { Theme } from '@material-ui/core/styles';
import { makeStyles } from '@material-ui/styles';

const useStyles = makeStyles((theme: Theme) => ({
  paper: {
    padding: theme.spacing(2.5),
    overflowX: 'auto',
    border: `1px solid ${theme.palette.divider}`,
    boxShadow: '0 10px 35px rgba(20, 29, 60, 0.08)',
    '& h3': {
      margin: theme.spacing(0, 0, 2),
      fontSize: '1.15rem',
      fontWeight: 600,
    },
  },
  editButton: {
    borderRadius: 8,
    color: theme.palette.primary.main,
  },
}));

export default useStyles;
