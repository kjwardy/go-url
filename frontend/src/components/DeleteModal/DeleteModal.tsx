import React, { useState } from 'react';
import { connect } from 'react-redux';
import axios from 'axios';
import {
  displayFlashError,
  displayFlashSuccess,
} from '../../redux/flash/actions';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogHeader,
  AlertDialogTitle,
} from '../ui/alert-dialog';

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
    <AlertDialog open onOpenChange={(open) => !open && !deleting && onClose()}>
      <AlertDialogContent data-e2e="delete-modal">
        <AlertDialogHeader>
          <AlertDialogTitle>{`Delete ${urlKey}?`}</AlertDialogTitle>
          <AlertDialogDescription>
            {`Are you sure you want to delete "${urlKey}"? This cannot be undone.`}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <div className="flex justify-end gap-2">
          <AlertDialogCancel data-e2e="delete-cancel" disabled={deleting}>
            Cancel
          </AlertDialogCancel>
          <AlertDialogAction
            className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            onSelect={(event) => {
              event.preventDefault();
              deleteUrl();
            }}
            data-e2e="delete-confirm"
            disabled={deleting}
          >
            Delete
          </AlertDialogAction>
        </div>
      </AlertDialogContent>
    </AlertDialog>
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
