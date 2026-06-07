import { cn } from "@/lib/utils";

type ChatTypingIndicatorProps = {
  className?: string;
};

export function ChatTypingIndicator({ className }: ChatTypingIndicatorProps) {
  return (
    <span
      role="status"
      aria-label="Assistant is typing"
      data-testid="chat-typing-indicator"
      className={cn("inline-flex items-center gap-1 align-middle", className)}
    >
      {[0, 1, 2].map((dot) => (
        <span
          key={dot}
          className="size-1.5 animate-bounce rounded-full bg-muted-foreground/60"
          style={{ animationDelay: `${dot * 0.16}s`, animationDuration: "1s" }}
        />
      ))}
    </span>
  );
}
