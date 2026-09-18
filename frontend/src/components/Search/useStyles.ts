import { alpha, Theme } from '@material-ui/core/styles';
import { makeStyles } from '@material-ui/styles';

const useStyles = makeStyles((theme: Theme) => ({
  search: {
    position: 'relative',
    width: '100%',
    marginLeft: 0,
    border: `1px solid ${alpha(theme.palette.common.white, 0.16)}`,
    borderRadius: 24,
    backgroundColor: alpha(theme.palette.common.white, 0.1),
    transition: theme.transitions.create(['background-color', 'border-color']),
    '&:hover, &:focus-within': {
      borderColor: alpha(theme.palette.common.white, 0.32),
      backgroundColor: alpha(theme.palette.common.white, 0.16),
    },
    [theme.breakpoints.up('sm')]: {
      marginLeft: theme.spacing(1),
      width: 'auto',
    },
  },
  searchIcon: {
    position: 'absolute',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    width: theme.spacing(6),
    height: '100%',
    pointerEvents: 'none',
  },
  inputRoot: {
    width: '100%',
    color: 'inherit',
  },
  inputInput: {
    width: '100%',
    padding: theme.spacing(1.25, 2, 1.25, 6),
    transition: theme.transitions.create('width'),
    [theme.breakpoints.up('sm')]: {
      width: 280,
      '&:focus': {
        width: 360,
      },
    },
  },
}));

export default useStyles;
