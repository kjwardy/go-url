import React, { useState, useEffect, useRef } from 'react';
import { connect } from 'react-redux';
import axios from 'axios';
import Button from '@material-ui/core/Button';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogContentText from '@material-ui/core/DialogContentText';
import DialogTitle from '@material-ui/core/DialogTitle';
import TextField from '@material-ui/core/TextField';
import {
  displayFlashError,
  displayFlashSuccess,
} from '../../redux/flash/actions';
import { urlCreated, urlUpdated } from '../../redux/search/actions';
import useStyles from './useStyles';

interface EditModalProps {
  edit?: Boolean;
  urlKey?: string;
  url?: string;
  onClose: () => void;
  onCreated?: (data: any) => void;
  displayFlashSuccess: (message: string) => void;
  displayFlashError: (message: string) => void;
  urlCreated: (data: any) => void;
  urlUpdated: (data: any) => void;
}

const EditModal: React.FC<EditModalProps> = ({
  edit,
  urlKey: initialKey = '',
  url: initialUrl = '',
  onClose,
  onCreated,
  displayFlashSuccess,
  displayFlashError,
  urlCreated,
  urlUpdated,
}) => {
  const [urlKey, setKey] = useState(initialKey);
  const [url, setUrl] = useState(initialUrl);
  const [query, submit] = useState<{ urlKey: string; url: string }>();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const submittedQuery = useRef<typeof query>();
  const classes = useStyles({});

  useEffect(() => {
    if (!query || submittedQuery.current === query) return;
    submittedQuery.current = query;
    setIsSubmitting(true);
    axios({
      method: edit ? 'put' : 'post',
      url: `/${encodeURIComponent(query.urlKey)}`,
      data: { url: query.url },
    })
      .then(({ data }: any) => {
        displayFlashSuccess(
          `Successfully set ${data.key} to ${data.url || data.alias}`,
        );
        // Update the displayed results without reloading the page
        if (edit) {
          urlUpdated(data);
        } else {
          urlCreated(data);
          if (onCreated) onCreated(data);
        }
        onClose();
      })
      .catch((err) => {
        setIsSubmitting(false);
        displayFlashError(err.response.data.message || err.response.data);
      });
  }, [
    query,
    displayFlashSuccess,
    displayFlashError,
    onClose,
    onCreated,
    edit,
    urlCreated,
    urlUpdated,
  ]);

  return (
    <Dialog open onClose={onClose} data-e2e="modal">
      <form
        onSubmit={(e) => {
          e.preventDefault();
          if (!isSubmitting) submit({ urlKey, url });
        }}
      >
        <DialogTitle>{edit ? `Edit ${urlKey}` : 'Add new URL'}</DialogTitle>
        <DialogContent>
          <DialogContentText>
            {edit
              ? `You are editing the link for "${urlKey}". Please remember that this will change the URL for everyone, so only do so if the URL is wrong.`
              : 'Enter key and URL to add new link'}
          </DialogContentText>
          {!edit && (
            <TextField
              id="key"
              label="Key"
              type="text"
              className={classes.textField}
              fullWidth
              autoComplete="off"
              onChange={(e) => setKey(e.target.value)}
              value={urlKey}
            />
          )}
          <TextField
            id="url"
            label="URL"
            type="text"
            className={classes.textField}
            fullWidth
            autoComplete="off"
            onChange={(e) => setUrl(e.target.value)}
            value={url}
          />
        </DialogContent>
        <DialogActions className={classes.actions}>
          <Button
            onClick={onClose}
            color="secondary"
            data-e2e="cancel"
            type="button"
            disabled={isSubmitting}
          >
            Cancel
          </Button>
          <Button
            type="submit"
            color="primary"
            data-e2e="submit"
            disabled={isSubmitting}
          >
            {edit ? 'Update' : 'Add'}
          </Button>
        </DialogActions>
      </form>
    </Dialog>
  );
};

const mapDispatch = {
  displayFlashSuccess,
  displayFlashError,
  urlCreated,
  urlUpdated,
};

export default connect(
  null,
  mapDispatch,
  // @ts-ignore
)(EditModal);
