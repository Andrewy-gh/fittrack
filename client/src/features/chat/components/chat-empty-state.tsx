import type { ReactNode } from "react";
import { cn } from "@/lib/utils";

type ChatEmptyStateProps = {
  heading: string;
  examples: string[];
  onSelectExample: (text: string) => void;
  composer: ReactNode;
  disabled?: boolean;
};

export function ChatEmptyState({
  heading,
  examples,
  onSelectExample,
  composer,
  disabled = false,
}: ChatEmptyStateProps) {
  return (
    <div className="flex min-h-[70vh] w-full flex-col items-center justify-center gap-6 px-1">
      <h1 className="text-center text-2xl font-semibold tracking-tight text-foreground sm:text-3xl">
        {heading}
      </h1>
      <div className="w-full">{composer}</div>
      {examples.length > 0 ? (
        <div className="flex flex-wrap items-center justify-center gap-2">
          {examples.map((example, index) => (
            <button
              key={example}
              type="button"
              disabled={disabled}
              onClick={() => onSelectExample(example)}
              className={cn(
                "rounded-full border bg-background px-4 py-2 text-sm text-muted-foreground transition-colors",
                "hover:bg-accent hover:text-accent-foreground disabled:cursor-not-allowed disabled:opacity-50",
                // One example on small screens, a second on larger screens.
                index >= 1 && "hidden sm:inline-flex",
              )}
            >
              {example}
            </button>
          ))}
        </div>
      ) : null}
    </div>
  );
}
