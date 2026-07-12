import { useEffect, useRef } from "react";
import type { FormEvent, KeyboardEvent } from "react";
import { ArrowUp, Square } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

type ChatComposerProps = {
  value: string;
  onChange: (value: string) => void;
  onSubmit: () => void;
  onStop?: () => void;
  disabled?: boolean;
  isSubmitting?: boolean;
  placeholder?: string;
  autoFocus?: boolean;
  className?: string;
};

export function ChatComposer({
  value,
  onChange,
  onSubmit,
  onStop,
  disabled = false,
  isSubmitting = false,
  placeholder,
  autoFocus = false,
  className,
}: ChatComposerProps) {
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const canSend = Boolean(value.trim()) && !disabled && !isSubmitting;

  useEffect(() => {
    const element = textareaRef.current;
    if (!element) {
      return;
    }
    element.style.height = "auto";
    element.style.height = `${Math.min(element.scrollHeight, 160)}px`;
  }, [value]);

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (canSend) {
      onSubmit();
    }
  }

  function handleKeyDown(event: KeyboardEvent<HTMLTextAreaElement>) {
    if (
      event.key === "Enter" &&
      !event.shiftKey &&
      !event.nativeEvent.isComposing
    ) {
      event.preventDefault();
      if (canSend) {
        onSubmit();
      }
    }
  }

  return (
    <form
      onSubmit={handleSubmit}
      className={cn("w-full", className)}
    >
      <div
        className={cn(
          "flex items-end gap-2 rounded-[1.75rem] border bg-muted/40 px-4 py-3 shadow-sm",
          "transition-colors focus-within:border-ring focus-within:bg-background",
          disabled && "opacity-70",
        )}
      >
        <textarea
          ref={textareaRef}
          value={value}
          onChange={(event) => onChange(event.target.value)}
          onKeyDown={handleKeyDown}
          placeholder={placeholder}
          disabled={disabled}
          // eslint-disable-next-line jsx-a11y/no-autofocus
          autoFocus={autoFocus}
          rows={1}
          className={cn(
            "max-h-40 min-h-[1.5rem] flex-1 resize-none bg-transparent py-1 text-base leading-6 md:text-sm",
            "outline-none placeholder:text-muted-foreground disabled:cursor-not-allowed",
          )}
        />
        <Button
          type={isSubmitting ? "button" : "submit"}
          size="icon"
          aria-label={isSubmitting ? "Stop response" : "Send"}
          onClick={isSubmitting ? onStop : undefined}
          disabled={isSubmitting ? !onStop : !canSend}
          className="size-9 shrink-0 rounded-full"
        >
          {isSubmitting ? (
            <Square className="size-4 fill-current" />
          ) : (
            <ArrowUp className="size-4" />
          )}
        </Button>
      </div>
    </form>
  );
}
