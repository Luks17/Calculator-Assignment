import { CircleAlert, X } from "lucide-react";
import { useEffect } from "react";

type CalculationErrorToastProps = {
  message: string;
  onDismiss: () => void;
};

export function CalculationErrorToast({
  message,
  onDismiss,
}: CalculationErrorToastProps) {
  useEffect(() => {
    const timeout = window.setTimeout(onDismiss, 6000);
    return () => window.clearTimeout(timeout);
  }, [message, onDismiss]);

  return (
    <div className="toast toast-end toast-bottom z-50 max-w-full">
      <div className="alert alert-error max-w-md shadow-lg" role="alert">
        <CircleAlert size={20} aria-hidden="true" />
        <span className="min-w-0 break-words">{message}</span>
        <button
          type="button"
          className="btn btn-ghost btn-square btn-sm"
          aria-label="Dismiss error"
          onClick={onDismiss}
        >
          <X size={18} aria-hidden="true" />
        </button>
      </div>
    </div>
  );
}
