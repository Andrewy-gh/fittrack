import { Check, Copy, RotateCcw } from "lucide-react";
import { useEffect, useState } from "react";
import { toast } from "sonner";
import { cn } from "@/lib/utils";

type ChatMessageActionsProps = {
  content: string;
  canRetry?: boolean;
  onRetry?: () => void;
  className?: string;
};

export function ChatMessageActions({
  content,
  canRetry = false,
  onRetry,
  className,
}: ChatMessageActionsProps) {
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    if (!copied) {
      return;
    }
    const timer = window.setTimeout(() => setCopied(false), 1500);
    return () => window.clearTimeout(timer);
  }, [copied]);

  async function handleCopy() {
    try {
      await navigator.clipboard?.writeText(content);
      setCopied(true);
    } catch {
      toast.error("Could not copy message");
    }
  }

  return (
    <div
      className={cn(
        "flex items-center gap-0.5 text-muted-foreground",
        className,
      )}
    >
      <ActionButton
        label={copied ? "Copied" : "Copy"}
        onClick={handleCopy}
      >
        {copied ? <Check className="size-4" /> : <Copy className="size-4" />}
      </ActionButton>
      {canRetry && onRetry ? (
        <ActionButton
          label="Retry"
          onClick={onRetry}
        >
          <RotateCcw className="size-4" />
        </ActionButton>
      ) : null}
    </div>
  );
}

function ActionButton({
  label,
  onClick,
  children,
}: {
  label: string;
  onClick: () => void;
  children: React.ReactNode;
}) {
  return (
    <button
      type="button"
      aria-label={label}
      title={label}
      onClick={onClick}
      className="inline-flex size-7 items-center justify-center rounded-md transition-colors hover:bg-accent hover:text-accent-foreground"
    >
      {children}
    </button>
  );
}
