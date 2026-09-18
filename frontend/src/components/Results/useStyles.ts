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
  url: {
    color: theme.palette.primary.main,
    textDecoration: 'none',
    fontWeight: 500,
    '&:hover': {
      textDecoration: 'underline',
    },
  },
  launchIcon: {
    width: 12,
    marginLeft: 3,
  },
  edit: {
    color: theme.palette.primary.main,
  },
  delete: {
    color: theme.palette.error.main,
  },
  actionIcon: {
    padding: theme.spacing(0.75),
    margin: theme.spacing(0, 0.25),
    borderRadius: 8,
  },
  tableRow: {
    height: 'initial',
    transition: theme.transitions.create('background-color'),
  },
  urlCell: {
    maxWidth: 300,
    overflowWrap: 'break-word',
    wordWrap: 'break-word',
    [theme.breakpoints.down('xs')]: {
      maxWidth: 100,
    },
  },
  urlReplace: {
    color: theme.palette.text.primary,
    fontWeight: 700,
  },
}));

export default useStyles;
