import React, { useEffect, useState } from 'react';
import axios from 'axios';
import Paper from '@material-ui/core/Paper';
import Tooltip from '@material-ui/core/Tooltip';
import Typography from '@material-ui/core/Typography';
import AssessmentIcon from '@material-ui/icons/Assessment';
import CheckCircleOutlineIcon from '@material-ui/icons/CheckCircleOutline';
import ErrorOutlineIcon from '@material-ui/icons/ErrorOutline';
import useStyles from './useStyles';

interface DailyQueries {
  date: string;
  successful: number;
  failed: number;
  total: number;
}

interface QueryMetrics {
  total_all_time: number;
  failed_all_time: number;
  success_percentage: number;
  total_last_seven_days: number;
  failed_last_seven_days: number;
  success_percentage_last_seven_days: number;
  daily_queries: DailyQueries[];
}

interface MetricsProps {
  displayFlashError: (message: string) => void;
}

const Metrics: React.FC<MetricsProps> = ({ displayFlashError }) => {
  const [metrics, setMetrics] = useState<QueryMetrics>();
  const classes = useStyles({});

  useEffect(() => {
    const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone;
    axios
      .get<QueryMetrics>('/api/metrics', { params: { timezone } })
      .then(({ data }) => setMetrics(data))
      .catch((err) =>
        displayFlashError(err.response.data.message || err.response.data),
      );
  }, [displayFlashError]);

  const dailyQueries = metrics?.daily_queries || [];
  const maxDailyQueries = Math.max(
    ...dailyQueries.map(({ total }) => total),
    1,
  );
  const values = [
    {
      label: 'Total queries',
      value: metrics?.total_all_time,
      recentValue: metrics?.total_last_seven_days,
      icon: <AssessmentIcon />,
      className: classes.total,
    },
    {
      label: 'Successful queries',
      value:
        typeof metrics?.success_percentage === 'number'
          ? `${metrics.success_percentage.toLocaleString()}%`
          : undefined,
      recentValue:
        typeof metrics?.success_percentage_last_seven_days === 'number'
          ? `${metrics.success_percentage_last_seven_days.toLocaleString()}%`
          : undefined,
      icon: <CheckCircleOutlineIcon />,
      className: classes.success,
    },
    {
      label: 'Failed queries',
      value: metrics?.failed_all_time,
      recentValue: metrics?.failed_last_seven_days,
      icon: <ErrorOutlineIcon />,
      className: classes.failed,
    },
  ];

  return (
    <Paper className={classes.paper} component="aside">
      <div className={classes.header}>
        <Typography variant="overline" component="p" color="textSecondary">
          Overview
        </Typography>
        <Typography variant="h6" component="h2">
          Query metrics
        </Typography>
      </div>
      <div className={classes.metrics}>
        {values.map(({ label, value, recentValue, icon, className }) => (
          <div className={`${classes.metric} ${className}`} key={label}>
            <div className={classes.icon}>{icon}</div>
            <div>
              <Typography variant="h4" component="p" className={classes.value}>
                {typeof value === 'number'
                  ? value.toLocaleString()
                  : value || '—'}
              </Typography>
              <Typography variant="body2" color="textSecondary">
                {label}
              </Typography>
              <Typography variant="caption" className={classes.recent}>
                <strong>
                  {typeof recentValue === 'number'
                    ? recentValue.toLocaleString()
                    : recentValue || '—'}
                </strong>{' '}
                in the last 7 days
              </Typography>
            </div>
          </div>
        ))}
      </div>
      <div className={classes.chartHeader}>
        <Typography variant="subtitle2">Queries by day</Typography>
        <div className={classes.legend}>
          <span className={classes.successKey}>Successful</span>
          <span className={classes.failedKey}>Failed</span>
        </div>
      </div>
      <div
        className={classes.chart}
        aria-label="Queries during the last seven days"
      >
        {dailyQueries.map((day) => {
          const date = new Date(`${day.date}T00:00:00`);
          const fullDate = date.toLocaleDateString(undefined, {
            month: 'short',
            day: 'numeric',
          });
          return (
            <Tooltip
              key={day.date}
              title={`${fullDate}: ${day.total.toLocaleString()} total (${day.successful.toLocaleString()} successful, ${day.failed.toLocaleString()} failed)`}
            >
              <div className={classes.day}>
                <div className={classes.barTrack}>
                  <div
                    className={classes.failedBar}
                    style={{
                      height: `${(day.failed / maxDailyQueries) * 100}%`,
                    }}
                  />
                  <div
                    className={classes.successBar}
                    style={{
                      height: `${(day.successful / maxDailyQueries) * 100}%`,
                    }}
                  />
                </div>
                <Typography variant="caption" color="textSecondary">
                  {date.toLocaleDateString(undefined, { weekday: 'narrow' })}
                </Typography>
              </div>
            </Tooltip>
          );
        })}
      </div>
    </Paper>
  );
};

export default Metrics;
