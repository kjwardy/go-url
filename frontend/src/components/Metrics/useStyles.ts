import { Theme } from '@material-ui/core/styles';
import { makeStyles } from '@material-ui/styles';

const useStyles = makeStyles((theme: Theme) => ({
  paper: {
    padding: theme.spacing(2.5),
    alignSelf: 'start',
    borderRadius: theme.spacing(1.5),
    overflow: 'hidden',
  },
  header: {
    marginBottom: theme.spacing(2),
    '& p': {
      lineHeight: 1.2,
      letterSpacing: 1.5,
    },
  },
  metrics: {
    display: 'grid',
    gap: theme.spacing(1.5),
  },
  metric: {
    display: 'flex',
    alignItems: 'center',
    gap: theme.spacing(2),
    padding: theme.spacing(2),
    border: `1px solid ${theme.palette.divider}`,
    borderRadius: theme.spacing(1),
    borderLeftWidth: 4,
    backgroundColor: theme.palette.background.default,
  },
  icon: {
    display: 'flex',
    padding: theme.spacing(1),
    borderRadius: '50%',
    backgroundColor: theme.palette.action.hover,
  },
  value: {
    lineHeight: 1.1,
    fontWeight: 600,
  },
  recent: {
    display: 'inline-block',
    marginTop: theme.spacing(1),
    padding: theme.spacing(0.4, 0.75),
    borderRadius: 6,
    color: theme.palette.text.secondary,
    backgroundColor: theme.palette.action.hover,
    whiteSpace: 'nowrap',
  },
  total: {
    borderLeftColor: theme.palette.primary.main,
    '& $icon': {
      color: theme.palette.primary.main,
    },
  },
  success: {
    borderLeftColor: theme.palette.success.main,
    '& $icon': {
      color: theme.palette.success.main,
    },
  },
  failed: {
    borderLeftColor: theme.palette.error.main,
    '& $icon': {
      color: theme.palette.error.main,
    },
  },
  chartHeader: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'flex-end',
    gap: theme.spacing(1),
    marginTop: theme.spacing(3),
    marginBottom: theme.spacing(1.5),
  },
  legend: {
    display: 'flex',
    gap: theme.spacing(1.5),
    color: theme.palette.text.secondary,
    fontSize: '0.65rem',
  },
  successKey: {
    '&::before': {
      content: '""',
      display: 'inline-block',
      width: 7,
      height: 7,
      marginRight: 4,
      borderRadius: 2,
      backgroundColor: theme.palette.success.main,
    },
  },
  failedKey: {
    '&::before': {
      content: '""',
      display: 'inline-block',
      width: 7,
      height: 7,
      marginRight: 4,
      borderRadius: 2,
      backgroundColor: theme.palette.error.main,
    },
  },
  chart: {
    display: 'grid',
    gridTemplateColumns: 'repeat(7, minmax(0, 1fr))',
    gap: theme.spacing(1),
    height: 145,
    paddingTop: theme.spacing(1),
    borderBottom: `1px solid ${theme.palette.divider}`,
  },
  day: {
    display: 'grid',
    gridTemplateRows: '1fr auto',
    gap: theme.spacing(0.5),
    minWidth: 0,
    textAlign: 'center',
  },
  barTrack: {
    display: 'flex',
    flexDirection: 'column',
    justifyContent: 'flex-end',
    alignSelf: 'stretch',
    minHeight: 0,
    borderRadius: theme.spacing(0.5, 0.5, 0, 0),
    overflow: 'hidden',
    backgroundColor: theme.palette.action.hover,
  },
  successBar: {
    flexShrink: 0,
    minHeight: 0,
    backgroundColor: theme.palette.success.main,
  },
  failedBar: {
    flexShrink: 0,
    minHeight: 0,
    backgroundColor: theme.palette.error.main,
  },
}));

export default useStyles;
