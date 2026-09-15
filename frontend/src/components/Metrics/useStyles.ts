import { Theme } from '@material-ui/core/styles';
import { makeStyles } from '@material-ui/styles';

const useStyles = makeStyles((theme: Theme) => ({
  paper: {
    padding: theme.spacing(2),
    alignSelf: 'start',
  },
  heading: {
    marginBottom: theme.spacing(1),
  },
  metric: {
    padding: theme.spacing(2, 0),
    borderBottom: `1px solid ${theme.palette.divider}`,
    '&:last-child': {
      borderBottom: 0,
      paddingBottom: 0,
    },
  },
}));

export default useStyles;
