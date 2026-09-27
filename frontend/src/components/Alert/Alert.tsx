import React, { useEffect } from 'react';
import { connect } from 'react-redux';
import { clearFlash } from '../../redux/flash/actions';
import { toast } from '../ui/toast';

export type Variant = 'success' | 'warning' | 'error' | 'info';

interface AlertProps {
  variant: Variant;
  message: React.ReactNode;
  clearFlash: () => void;
}

const Alert = ({ variant, message, clearFlash }: AlertProps) => {
  useEffect(() => {
    toast.add({
      id: 'flash-message',
      title: message,
      type: variant,
      timeout: 6000,
      priority: variant === 'error' ? 'high' : 'low',
      onClose: clearFlash,
    });
  }, [clearFlash, message, variant]);

  return null;
};

const mapDispatch = {
  clearFlash,
};

export default connect(null, mapDispatch)(Alert);
