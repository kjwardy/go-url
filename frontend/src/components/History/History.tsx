import React, { useEffect, useState } from 'react';
import axios from 'axios';
import FormControl from '@material-ui/core/FormControl';
import InputLabel from '@material-ui/core/InputLabel';
import MenuItem from '@material-ui/core/MenuItem';
import Paper from '@material-ui/core/Paper';
import Select from '@material-ui/core/Select';
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableHead from '@material-ui/core/TableHead';
import TableRow from '@material-ui/core/TableRow';
import Tooltip from '@material-ui/core/Tooltip';
import CheckIcon from '@material-ui/icons/Check';
import CloseIcon from '@material-ui/icons/Close';
import useStyles from './useStyles';

type HistoryLimit = 25 | 50 | 100;

interface HistoryEntry {
  id: number;
  url_key: string;
  queried_at: string;
  successful: boolean;
}

interface HistoryProps {
  displayFlashError: (message: string) => void;
}

const History: React.FC<HistoryProps> = ({ displayFlashError }) => {
  const [limit, setLimit] = useState<HistoryLimit>(25);
  const [history, setHistory] = useState<HistoryEntry[]>();
  const classes = useStyles({});

  useEffect(() => {
    axios
      .get<HistoryEntry[]>('/api/history', { params: { limit } })
      .then(({ data }) => setHistory(data))
      .catch((err) =>
        displayFlashError(err.response.data.message || err.response.data),
      );
  }, [limit, displayFlashError]);

  return (
    <Paper className={classes.paper}>
      <div className={classes.header}>
        <h3>Query History</h3>
        <FormControl variant="outlined" size="small">
          <InputLabel id="history-limit-label">Records</InputLabel>
          <Select
            labelId="history-limit-label"
            id="history-limit"
            value={limit}
            onChange={(event) =>
              setLimit(Number(event.target.value) as HistoryLimit)
            }
            label="Records"
          >
            <MenuItem value={25}>25</MenuItem>
            <MenuItem value={50}>50</MenuItem>
            <MenuItem value={100}>100</MenuItem>
          </Select>
        </FormControl>
      </div>
      {history && history.length === 0 ? (
        <p>No query history found - go nuts!</p>
      ) : (
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>Timestamp</TableCell>
              <TableCell>Query</TableCell>
              <TableCell align="center">Successful</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {(history || []).map((entry) => (
              <TableRow key={entry.id}>
                <TableCell>
                  {new Date(entry.queried_at).toLocaleString()}
                </TableCell>
                <TableCell>{entry.url_key}</TableCell>
                <TableCell align="center">
                  {entry.successful ? (
                    <Tooltip title="Resolved successfully">
                      <CheckIcon
                        className={classes.successful}
                        aria-label="Resolved successfully"
                      />
                    </Tooltip>
                  ) : (
                    <Tooltip title="Did not resolve">
                      <CloseIcon
                        className={classes.unsuccessful}
                        aria-label="Did not resolve"
                      />
                    </Tooltip>
                  )}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}
    </Paper>
  );
};

export default History;
