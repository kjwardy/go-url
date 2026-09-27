import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { Activity, CheckCircle2, CircleAlert } from 'lucide-react';
import { Card } from '../ui/card';
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
      className: 'border-l-primary text-primary',
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
      className: 'border-l-emerald-600 text-emerald-600',
    },
    {
      label: 'Failed queries',
      value: metrics?.failed_all_time,
      recentValue: metrics?.failed_last_seven_days,
      icon: <CircleAlert />,
      className: 'border-l-destructive text-destructive',
    },
  ];

  return (
    <Card className="self-start overflow-hidden p-5">
      <aside>
        <div className="mb-4">
          <p className="text-xs uppercase leading-tight tracking-[1.5px] text-muted-foreground">
            Overview
          </p>
          <h2 className="mt-1 text-lg font-semibold">Query metrics</h2>
        </div>
        <div className="grid gap-3">
          {values.map(({ label, value, recentValue, icon, className }) => (
            <div
              className={`flex items-center gap-4 rounded-lg border border-border border-l-4 bg-background p-4 ${className}`}
              key={label}
            >
              <div className="flex rounded-full bg-muted p-2">{icon}</div>
              <div>
                <p className="text-2xl font-semibold leading-tight text-foreground">
                  {typeof value === 'number'
                    ? value.toLocaleString()
                    : value || '—'}
                </p>
                <p className="text-sm text-muted-foreground">{label}</p>
                <p className="mt-2 inline-block whitespace-nowrap rounded-md bg-muted px-2 py-1 text-xs text-muted-foreground">
                  <strong>
                    {typeof recentValue === 'number'
                      ? recentValue.toLocaleString()
                      : recentValue || '—'}
                  </strong>{' '}
                  in the last 7 days
                </p>
              </div>
            </div>
          ))}
        </div>
        <div className="mb-3 mt-6 flex items-end justify-between gap-2">
          <h3 className="text-sm font-medium">Queries by day</h3>
          <div className="flex gap-3 text-[0.65rem] text-muted-foreground">
            <span className="before:mr-1 before:inline-block before:h-[7px] before:w-[7px] before:rounded-sm before:bg-emerald-600 before:content-['']">
              Successful
            </span>
            <span className="before:mr-1 before:inline-block before:h-[7px] before:w-[7px] before:rounded-sm before:bg-destructive before:content-['']">
              Failed
            </span>
          </div>
        </div>
        <div
          className="grid h-[145px] grid-cols-7 gap-2 border-b border-border pt-2"
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
                <TooltipTrigger asChild>
                  <div className="grid min-w-0 grid-rows-[1fr_auto] gap-1 text-center">
                    <div className="flex min-h-0 flex-col justify-end self-stretch overflow-hidden rounded-t bg-muted">
                      <div
                        className="min-h-0 shrink-0 bg-destructive"
                        style={{
                          height: `${(day.failed / maxDailyQueries) * 100}%`,
                        }}
                      />
                      <div
                        className="min-h-0 shrink-0 bg-emerald-600"
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
                  </div>
                </TooltipTrigger>
                <TooltipContent>
                  {`${fullDate}: ${day.total.toLocaleString()} total (${day.successful.toLocaleString()} successful, ${day.failed.toLocaleString()} failed)`}
                </TooltipContent>
              </Tooltip>
            );
          })}
        </div>
      </aside>
    </Card>
  );
};

export default Metrics;
