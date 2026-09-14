import React, { useState } from 'react';
import { connect } from 'react-redux';
import axios from 'axios';
import Button from '@material-ui/core/Button';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogContentText from '@material-ui/core/DialogContentText';
import DialogTitle from '@material-ui/core/DialogTitle';
import {
  displayFlashError,
  displayFlashSuccess,
} from '../../redux/flash/actions';

interface DeleteModalOwnProps {
  urlKey: string;
  onClose: () => void;
  onDeleted: () => void;
}

interface DeleteModalDispatchProps {
  displayFlashSuccess: (message: string) => void;
  displayFlashError: (message: string) => void;
}

type DeleteModalProps = DeleteModalOwnProps & DeleteModalDispatchProps;

const DeleteModal: React.FC<DeleteModalProps> = ({
  urlKey,
  onClose,
  onDeleted,
  displayFlashSuccess,
  displayFlashError,
}) => {
  const [deleting, setDeleting] = useState(false);

  // Disable both actions while deleting to prevent duplicate requests
  const deleteUrl = () => {
    setDeleting(true);
    axios
      .delete(`/${encodeURIComponent(urlKey)}`)
      .then(() => {
        displayFlashSuccess(`Successfully deleted ${urlKey}`);
        onDeleted();
        onClose();
      })
      .catch((err) => {
        setDeleting(false);
        displayFlashError(err.response.data.message || err.response.data);
      });
  };

  return (
    <Dialog open onClose={onClose} data-e2e="delete-modal">
      <DialogTitle>{`Delete ${urlKey}?`}</DialogTitle>
      <DialogContent>
        <DialogContentText>
          {`Are you sure you want to delete "${urlKey}"? This cannot be undone.`}
        </DialogContentText>
      </DialogContent>
      <DialogActions>
        <Button
          onClick={onClose}
          color="primary"
          data-e2e="delete-cancel"
          disabled={deleting}
        >
          Cancel
        </Button>
        <Button
          onClick={deleteUrl}
          color="secondary"
          data-e2e="delete-confirm"
          disabled={deleting}
        >
          Delete
        </Button>
      </DialogActions>
    </Dialog>
  );
};

const mapDispatch = {
  displayFlashSuccess,
  displayFlashError,
};

export default connect<DeleteModalDispatchProps, {}, DeleteModalOwnProps, {}>(
  null,
  mapDispatch,
)(DeleteModal);
