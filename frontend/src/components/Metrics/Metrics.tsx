import React, { useEffect, useState } from 'react';
import axios from 'axios';
import Paper from '@material-ui/core/Paper';
import Typography from '@material-ui/core/Typography';
import useStyles from './useStyles';

interface QueryMetrics {
  total_all_time: number;
  total_today: number;
  failed_all_time: number;
  failed_today: number;
}

interface MetricsProps {
  displayFlashError: (message: string) => void;
}

const Metrics: React.FC<MetricsProps> = ({ displayFlashError }) => {
  const [metrics, setMetrics] = useState<QueryMetrics>();
  const classes = useStyles({});

  useEffect(() => {
    axios
      .get<QueryMetrics>('/api/metrics')
      .then(({ data }) => setMetrics(data))
      .catch((err) =>
        displayFlashError(err.response.data.message || err.response.data),
      );
  }, [displayFlashError]);

  const values: Array<[string, number | undefined]> = [
    ['Total queries - all time', metrics?.total_all_time],
    ['Total queries - today', metrics?.total_today],
    ['Failed queries - all time', metrics?.failed_all_time],
    ['Failed queries - today', metrics?.failed_today],
  ];

  return (
    <Paper className={classes.paper} component="aside">
      <Typography variant="h6" component="h2" className={classes.heading}>
        Metrics
      </Typography>
      {values.map(([label, value]) => (
        <div className={classes.metric} key={label}>
          <Typography variant="h4" component="p">
            {typeof value === 'number' ? value.toLocaleString() : '—'}
          </Typography>
          <Typography variant="body2" color="textSecondary">
            {label}
          </Typography>
        </div>
      ))}
    </Paper>
  );
};

export default Metrics;
