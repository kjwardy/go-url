import React, { useState, useEffect, useRef } from 'react';
import { connect } from 'react-redux';
import axios from 'axios';
import {
  displayFlashError,
  displayFlashSuccess,
} from '../../redux/flash/actions';
import { urlCreated, urlUpdated } from '../../redux/search/actions';
import { Button } from '../ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '../ui/dialog';
import { Input } from '../ui/input';

interface EditModalProps {
  edit?: boolean;
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
    <Dialog open onOpenChange={(open) => !open && onClose()}>
      <DialogContent data-e2e="modal">
        <form
          className="space-y-5"
          onSubmit={(e) => {
            e.preventDefault();
            if (!isSubmitting) submit({ urlKey, url });
          }}
        >
          <DialogHeader>
            <DialogTitle>{edit ? `Edit ${urlKey}` : 'Add new URL'}</DialogTitle>
            <DialogDescription>
              {edit
                ? `You are editing the link for "${urlKey}". Please remember that this will change the URL for everyone, so only do so if the URL is wrong.`
                : 'Enter key and URL to add new link'}
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            {!edit && (
              <div className="space-y-2">
                <label htmlFor="key" className="text-sm font-medium">
                  Key
                </label>
                <Input
                  id="key"
                  type="text"
                  autoComplete="off"
                  onChange={(e) => setKey(e.target.value)}
                  value={urlKey}
                />
              </div>
            )}
            <div className="space-y-2">
              <label htmlFor="url" className="text-sm font-medium">
                URL
              </label>
              <Input
                id="url"
                type="text"
                autoComplete="off"
                onChange={(e) => setUrl(e.target.value)}
                value={url}
              />
            </div>
          </div>
          <div className="flex justify-end gap-2 pt-1">
            <Button
              onClick={onClose}
              variant="outline"
              data-e2e="cancel"
              type="button"
              disabled={isSubmitting}
            >
              Cancel
            </Button>
            <Button type="submit" data-e2e="submit" disabled={isSubmitting}>
              {edit ? 'Update' : 'Add'}
            </Button>
          </div>
        </form>
      </DialogContent>
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
