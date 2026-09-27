import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { Activity, CheckCircle2, CircleAlert } from 'lucide-react';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '../ui/card';
import { Badge } from '../ui/badge';
import { Tooltip, TooltipContent, TooltipTrigger } from '../ui/tooltip';

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
      icon: <Activity />,
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
      icon: <CheckCircle2 />,
    },
    {
      label: 'Failed queries',
      value: metrics?.failed_all_time,
      recentValue: metrics?.failed_last_seven_days,
      icon: <CircleAlert />,
    },
  ];

  return (
    <Card className="self-start">
      <aside>
        <CardHeader>
          <CardTitle>Overview</CardTitle>
          <CardDescription>Query metrics</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid gap-3">
            {values.map(({ label, value, recentValue, icon }) => (
              <div
                className="flex items-center gap-4 rounded-lg border bg-background p-4"
                key={label}
              >
                <div className="flex rounded-full bg-muted p-2 text-muted-foreground">
                  {icon}
                </div>
                <div>
                  <p className="text-2xl font-semibold leading-tight text-foreground">
                    {typeof value === 'number'
                      ? value.toLocaleString()
                      : value || '—'}
                  </p>
                  <p className="text-sm text-muted-foreground">{label}</p>
                  <Badge variant="secondary" className="mt-2">
                    <strong>
                      {typeof recentValue === 'number'
                        ? recentValue.toLocaleString()
                        : recentValue || '—'}
                    </strong>{' '}
                    in the last 7 days
                  </Badge>
                </div>
              </div>
            ))}
          </div>
          <div className="mb-3 mt-6 flex items-end justify-between gap-2">
            <h3 className="text-sm font-medium">Queries by day</h3>
            <div className="flex gap-3 text-xs text-muted-foreground">
              <span className="inline-flex items-center gap-1">
                <span className="h-2 w-2 rounded-sm bg-secondary" />
                Successful
              </span>
              <span className="inline-flex items-center gap-1">
                <span className="h-2 w-2 rounded-sm bg-destructive" />
                Failed
              </span>
            </div>
          </div>
          <div
            className="grid h-36 grid-cols-7 gap-2 border-b pt-2"
            aria-label="Queries during the last seven days"
            role="img"
          >
            {dailyQueries.map((day) => {
              const date = new Date(`${day.date}T00:00:00`);
              const fullDate = date.toLocaleDateString(undefined, {
                month: 'short',
                day: 'numeric',
              });
              return (
                <Tooltip key={day.date}>
                  <TooltipTrigger
                    render={
                      <div className="grid min-w-0 grid-rows-[1fr_auto] gap-1 text-center" />
                    }
                  >
                    <div className="flex min-h-0 flex-col justify-end self-stretch overflow-hidden rounded-t bg-muted">
                      <div
                        className="min-h-0 shrink-0 bg-destructive"
                        style={{
                          height: `${(day.failed / maxDailyQueries) * 100}%`,
                        }}
                      />
                      <div
                        className="min-h-0 shrink-0 bg-secondary"
                        style={{
                          height: `${
                            (day.successful / maxDailyQueries) * 100
                          }%`,
                        }}
                      />
                    </div>
                    <span className="text-xs text-muted-foreground">
                      {date.toLocaleDateString(undefined, {
                        weekday: 'narrow',
                      })}
                    </span>
                  </TooltipTrigger>
                  <TooltipContent>
                    {`${fullDate}: ${day.total.toLocaleString()} total (${day.successful.toLocaleString()} successful, ${day.failed.toLocaleString()} failed)`}
                  </TooltipContent>
                </Tooltip>
              );
            })}
          </div>
        </CardContent>
      </aside>
    </Card>
  );
};

export default Metrics;
