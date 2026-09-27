import React from 'react';
import { connect } from 'react-redux';
import { CheckCircle2, CircleAlert, Info, TriangleAlert } from 'lucide-react';
import { clearFlash } from '../../redux/flash/actions';
import { Toast, ToastClose, ToastTitle } from '../ui/toast';

export type Variant = 'success' | 'warning' | 'error' | 'info';

interface AlertProps {
  variant: Variant;
  message: React.ReactNode;
  clearFlash: () => void;
}

const alertVariant = {
  success: 'success',
  warning: 'warning',
  error: 'destructive',
  info: 'info',
} as const;

const variantIcon = {
  success: CheckCircle2,
  warning: TriangleAlert,
  error: CircleAlert,
  info: Info,
};

const Alert = ({ variant, message, clearFlash }: AlertProps) => {
  const Icon = variantIcon[variant];
  return (
    <Toast
      data-e2e="alert"
      variant={alertVariant[variant]}
      open
      onOpenChange={(open) => !open && clearFlash()}
    >
      <Icon aria-hidden="true" className="h-5 w-5" />
      <ToastTitle>{message}</ToastTitle>
      <ToastClose aria-label="Close alert" />
    </Toast>
  );
};

const mapDispatch = {
  clearFlash,
};

export default connect(null, mapDispatch)(Alert);
