import type { ReactNode } from "react";
import { ChevronDown, ChevronRight, Plus, Search } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Card, CardContent } from "@/components/ui/card";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";

/** Consistent search field for custom and catalog exercise lists. */
export function ExerciseListSearch({
  label,
  placeholder,
  value,
  onChange,
  disabled = false,
}: {
  label: string;
  placeholder: string;
  value: string;
  onChange: (value: string) => void;
  disabled?: boolean;
}) {
  return (
    <div className="relative">
      <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground w-4 h-4" />
      <Input
        autoFocus
        aria-label={label}
        placeholder={placeholder}
        value={value}
        onChange={(event) => onChange(event.target.value)}
        disabled={disabled}
        className="pl-10 text-base"
      />
    </div>
  );
}

const rowActionClass =
  "flex min-w-0 flex-1 items-center justify-between gap-3 p-4 text-left transition-colors hover:bg-gray-100/50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-inset disabled:opacity-50";

/** Shared exercise selection list; optional details expand independently of selection. */
export function ExerciseList<Entry extends { name: string }>({
  entries,
  onSelect,
  emptyMessage,
  renderDetails,
  disabled = false,
  showCount = false,
}: {
  entries: readonly Entry[];
  onSelect: (entry: Entry) => void;
  emptyMessage: ReactNode;
  renderDetails?: (entry: Entry) => ReactNode;
  disabled?: boolean;
  showCount?: boolean;
}) {
  return (
    <>
      <Card className="py-0">
        <CardContent className="p-0">
          {entries.length === 0 && (
            <div className="px-4 py-8 text-center text-wrap text-muted-foreground">
              {emptyMessage}
            </div>
          )}
          {entries.map((entry) =>
            renderDetails ? (
              <Collapsible
                key={entry.name}
                className="border-b border-border last:border-b-0"
              >
                <div className="flex items-stretch">
                  <CollapsibleTrigger
                    disabled={disabled}
                    aria-label={`Details for ${entry.name}`}
                    className={`group ${rowActionClass}`}
                  >
                    <span className="font-semibold md:text-sm">
                      {entry.name}
                    </span>
                    <ChevronDown className="w-5 h-5 shrink-0 text-muted-foreground transition-transform group-data-[state=open]:rotate-180" />
                  </CollapsibleTrigger>
                  <button
                    type="button"
                    aria-label={entry.name}
                    title={`Select ${entry.name}`}
                    disabled={disabled}
                    className="shrink-0 px-4 text-primary transition-colors hover:bg-gray-100/50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-inset disabled:opacity-50"
                    onClick={() => onSelect(entry)}
                  >
                    <Plus className="w-5 h-5" />
                  </button>
                </div>
                <CollapsibleContent className="space-y-1 px-4 pb-4 text-sm text-muted-foreground">
                  {renderDetails(entry)}
                </CollapsibleContent>
              </Collapsible>
            ) : (
              <button
                key={entry.name}
                type="button"
                disabled={disabled}
                className={`w-full border-b border-border last:border-b-0 ${rowActionClass}`}
                onClick={() => onSelect(entry)}
              >
                <span className="font-semibold md:text-sm">{entry.name}</span>
                <ChevronRight className="w-5 h-5 shrink-0 text-muted-foreground" />
              </button>
            ),
          )}
        </CardContent>
      </Card>
      {showCount && (
        <p className="text-center text-sm text-muted-foreground">
          {entries.length} exercise{entries.length !== 1 ? "s" : ""} found
        </p>
      )}
    </>
  );
}
