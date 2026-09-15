import React, { useEffect, useState } from 'react';
import axios from 'axios';
import Paper from '@material-ui/core/Paper';
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableHead from '@material-ui/core/TableHead';
import TableRow from '@material-ui/core/TableRow';
import IconButton from '@material-ui/core/IconButton';
import EditIcon from '@material-ui/icons/Edit';
import useStyles from './useStyles';

interface MostWantedEntry {
  query: string;
  views: number;
}

interface MostWantedProps {
  displayFlashError: (message: string) => void;
}

const MostWanted: React.FC<MostWantedProps> = ({ displayFlashError }) => {
  const [mostWanted, setMostWanted] = useState<MostWantedEntry[]>();
  const classes = useStyles({});

  useEffect(() => {
    axios
      .get<MostWantedEntry[]>('/api/most-wanted')
      .then(({ data }) => setMostWanted(data))
      .catch((err) =>
        displayFlashError(err.response.data.message || err.response.data),
      );
  }, [displayFlashError]);

  return (
    <Paper className={classes.paper}>
      <h3>Most Wanted</h3>
      {mostWanted && mostWanted.length === 0 ? (
        <p>No unresolved queries found.</p>
      ) : (
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>Query</TableCell>
              <TableCell align="right">Views</TableCell>
              <TableCell align="center">Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {(mostWanted || []).map((entry) => (
              <TableRow key={entry.query}>
                <TableCell>{entry.query}</TableCell>
                <TableCell align="right">{entry.views}</TableCell>
                <TableCell align="center">
                  <IconButton
                    aria-label="Edit"
                    className={classes.editButton}
                    disabled
                  >
                    <EditIcon />
                  </IconButton>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}
    </Paper>
  );
};

export default MostWanted;
