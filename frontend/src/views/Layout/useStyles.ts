import { Theme } from '@material-ui/core/styles';
import { makeStyles } from '@material-ui/styles';

const useStyles = makeStyles((theme: Theme) => ({
  root: {
    minHeight: '100vh',
    color: theme.palette.text.primary,
    backgroundColor: theme.palette.background.default,
    backgroundImage:
      theme.palette.type === 'light'
        ? 'radial-gradient(circle at 85% 0%, rgba(64, 84, 178, 0.08), transparent 28%)'
        : 'radial-gradient(circle at 85% 0%, rgba(142, 162, 255, 0.08), transparent 28%)',
  },
  button: {
    position: 'fixed',
    right: theme.spacing(3),
    bottom: theme.spacing(3),
  },
}));

export default useStyles;
