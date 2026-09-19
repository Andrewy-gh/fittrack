import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import type { ExercisecatalogEntry } from "@/client";
import { getExerciseCatalogQueryOptions } from "@/client/@tanstack/react-query.gen";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";

export function ExerciseCatalogPicker({
  onSelect,
  disabled = false,
}: {
  onSelect: (entry: ExercisecatalogEntry) => void;
  disabled?: boolean;
}) {
  const [search, setSearch] = useState("");
  const catalog = useQuery(getExerciseCatalogQueryOptions());
  const entries =
    catalog.data?.filter((entry) =>
      [
        entry.name,
        entry.equipment,
        ...entry.primaryMuscles,
        ...entry.secondaryMuscles,
      ]
        .join(" ")
        .toLowerCase()
        .includes(search.trim().toLowerCase()),
    ) ?? [];

  return (
    <div className="space-y-3">
      <Input
        aria-label="Search catalog"
        placeholder="Search name, equipment or muscle"
        value={search}
        onChange={(event) => setSearch(event.target.value)}
        disabled={disabled}
      />
      {catalog.isPending && <p role="status">Loading catalog...</p>}
      {catalog.isError && (
        <div role="alert">
          <p>Could not load the catalog.</p>
          <Button
            type="button"
            variant="outline"
            onClick={() => catalog.refetch()}
          >
            Retry
          </Button>
        </div>
      )}
      {catalog.isSuccess && entries.length === 0 && (
        <p>No catalog matches. You can keep a custom exercise unclassified.</p>
      )}
      <div className="max-h-72 overflow-y-auto divide-y rounded-md border">
        {entries.map((entry) => (
          <button
            key={entry.id}
            aria-label={entry.name}
            type="button"
            disabled={disabled}
            className="w-full p-3 text-left hover:bg-accent focus-visible:bg-accent disabled:opacity-50"
            onClick={() => onSelect(entry)}
          >
            <span className="block font-medium">{entry.name}</span>
            <span className="block text-sm text-muted-foreground">
              Equipment: {entry.equipment}
            </span>
            <span className="block text-sm text-muted-foreground">
              Primary: {entry.primaryMuscles.join(", ")}
            </span>
            <span className="block text-sm text-muted-foreground">
              Secondary: {entry.secondaryMuscles.join(", ") || "None listed"}
            </span>
          </button>
        ))}
      </div>
    </div>
  );
}
