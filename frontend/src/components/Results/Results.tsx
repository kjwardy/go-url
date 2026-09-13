import React, { useState, useCallback } from 'react';
import DeleteIcon from '@material-ui/icons/Delete';
import EditIcon from '@material-ui/icons/Edit';
import IconButton from '@material-ui/core/IconButton';
import LaunchIcon from '@material-ui/icons/Launch';
import Paper from '@material-ui/core/Paper';
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableHead from '@material-ui/core/TableHead';
import TableRow from '@material-ui/core/TableRow';

import DeleteModal from '../DeleteModal';
import EditModal from '../EditModal';
import useStyles from './useStyles';

interface IResult {
  key: string;
  url: string;
  alias: string[];
  views: number;
}

interface ResultsProps {
  data: IResult[];
  title: string;
}

const Results: React.FC<ResultsProps> = ({ data, title }) => {
  const [selected, setSelected] = useState<IResult | null>(null);
  const [deleteSelected, setDeleteSelected] = useState<IResult | null>(null);
  // Track deleted keys locally so rows disappear without reloading the page
  const [deletedKeys, setDeletedKeys] = useState<string[]>([]);
  const clearSelected = useCallback(() => setSelected(null), []);
  const clearDeleteSelected = useCallback(() => setDeleteSelected(null), []);
  const classes = useStyles({});
  const results = data.filter((result) => !deletedKeys.includes(result.key));

  const getFormattedUrl = (url: string) => {
    const regex = /({{\$\d+}})/g;
    const parts = url.split(regex);
    return parts.map((part, i) =>
      part.match(regex) ? (
        <span key={i} className={classes.urlReplace}>
          {part}
        </span>
      ) : (
        part
      ),
    );
  };
  return (
    <div>
      {selected && (
        <EditModal
          edit
          urlKey={selected.key}
          url={selected.url || selected.alias.join(',')}
          onClose={clearSelected}
        />
      )}
      {deleteSelected && (
        <DeleteModal
          urlKey={deleteSelected.key}
          onClose={clearDeleteSelected}
          onDeleted={() =>
            setDeletedKeys((keys) => [...keys, deleteSelected.key])
          }
        />
      )}

      <Paper className={classes.paper} data-e2e={title}>
        <h3>{title}</h3>
        {!results.length ? (
          <p>No results found. Help others by adding it.</p>
        ) : (
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>Key</TableCell>
                <TableCell>Url</TableCell>
                <TableCell align="right">Views</TableCell>
                <TableCell align="right">Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {results.map((r) => (
                <TableRow key={r.key} className={classes.tableRow}>
                  <TableCell>{r.key}</TableCell>
                  <TableCell className={classes.urlCell}>
                    {r.alias && r.alias.length ? (
                      r.alias.map((alias) => (
                        <a
                          key={alias}
                          className={classes.url}
                          href={`/${alias}`}
                        >
                          {alias}
                          <LaunchIcon className={classes.launchIcon} />
                        </a>
                      ))
                    ) : (
                      <a className={classes.url} href={`/${r.key}`}>
                        {getFormattedUrl(r.url)}
                        <LaunchIcon className={classes.launchIcon} />
                      </a>
                    )}
                  </TableCell>
                  <TableCell align="right">{r.views}</TableCell>
                  <TableCell align="right">
                    <IconButton
                      className={classes.actionIcon}
                      onClick={() => setSelected(r)}
                      aria-label={`Edit ${r.key}`}
                      data-e2e="edit"
                    >
                      <EditIcon className={classes.edit} />
                    </IconButton>
                    <IconButton
                      className={classes.actionIcon}
                      onClick={() => setDeleteSelected(r)}
                      aria-label={`Delete ${r.key}`}
                      data-e2e="delete"
                    >
                      <DeleteIcon className={classes.delete} />
                    </IconButton>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </Paper>
    </div>
  );
};

export default Results;
