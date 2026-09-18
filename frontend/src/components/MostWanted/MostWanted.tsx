import React, { useCallback, useEffect, useState } from 'react';
import axios from 'axios';
import Paper from '@material-ui/core/Paper';
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableHead from '@material-ui/core/TableHead';
import TableRow from '@material-ui/core/TableRow';
import IconButton from '@material-ui/core/IconButton';
import EditIcon from '@material-ui/icons/Edit';
import EditModal from '../EditModal';
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
  const [selected, setSelected] = useState<MostWantedEntry>();
  const classes = useStyles({});
  const fetchMostWanted = useCallback(() => {
    axios
      .get<MostWantedEntry[]>('/api/most-wanted')
      .then(({ data }) => setMostWanted(data))
      .catch((err) =>
        displayFlashError(err.response.data.message || err.response.data),
      );
  }, [displayFlashError]);

  useEffect(() => {
    fetchMostWanted();
  }, [fetchMostWanted]);

  return (
    <Paper className={classes.paper}>
      {selected && (
        <EditModal
          urlKey={selected.query}
          onClose={() => setSelected(undefined)}
          onCreated={() => {
            setMostWanted((entries) =>
              entries?.filter(({ query }) => query !== selected.query),
            );
          }}
        />
      )}
      <h3>Most Wanted</h3>
      {mostWanted && mostWanted.length === 0 ? (
        <p>No unresolved queries found - go nuts!</p>
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
                    aria-label={`Add URL for ${entry.query}`}
                    className={classes.editButton}
                    onClick={() => setSelected(entry)}
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
