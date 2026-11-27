import { Theme } from '@material-ui/core/styles';
import { makeStyles } from '@material-ui/styles';

const useStyles = makeStyles((theme: Theme) => ({
  root: {
    display: 'flex',
    flexDirection: 'column',
    minHeight: '100vh',
    paddingBottom: theme.spacing(7),
    color: theme.palette.text.primary,
    backgroundColor: theme.palette.background.default,
    backgroundImage:
      theme.palette.type === 'light'
        ? 'radial-gradient(circle at 85% 0%, rgba(64, 84, 178, 0.08), transparent 28%)'
        : 'radial-gradient(circle at 85% 0%, rgba(142, 162, 255, 0.08), transparent 28%)',
  },
  footer: {
    position: 'fixed',
    right: 0,
    bottom: 0,
    left: 0,
    zIndex: theme.zIndex.appBar,
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'flex-end',
    gap: theme.spacing(1),
    padding: theme.spacing(2, 3),
    borderTop: `1px solid ${theme.palette.divider}`,
    backgroundColor: theme.palette.background.default,
    color: theme.palette.text.secondary,
    fontSize: '0.75rem',
  },
  githubLink: {
    display: 'inline-flex',
    alignItems: 'center',
    gap: theme.spacing(0.5),
    color: 'inherit',
    textDecoration: 'none',
    '&:hover': {
      textDecoration: 'underline',
    },
  },
  githubIcon: {
    width: 16,
    height: 16,
  },
  button: {
    position: 'fixed',
    right: theme.spacing(3),
    bottom: theme.spacing(9),
  },
}));

export default useStyles;
