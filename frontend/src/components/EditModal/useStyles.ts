import { Theme } from '@material-ui/core/styles';
import { makeStyles } from '@material-ui/styles';

const useStyles = makeStyles((theme: Theme) => ({
  textField: {
    marginTop: theme.spacing(2),
  },
  actions: {
    marginTop: theme.spacing(2.5),
    gap: theme.spacing(1),
  },
}));

export default useStyles;
