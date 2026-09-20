import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import type { ExercisecatalogEntry } from "@/client";
import { getExerciseCatalogQueryOptions } from "@/client/@tanstack/react-query.gen";
import { Button } from "@/components/ui/button";
import { ExerciseList, ExerciseListSearch } from "./exercise-selection-list";

/** Search reviewed exercises, inspect their metadata, and explicitly select one. */
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
    <div className="space-y-6">
      <ExerciseListSearch
        label="Search catalog"
        placeholder="Search name, equipment or muscle"
        value={search}
        onChange={setSearch}
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
      {catalog.isSuccess && (
        <ExerciseList
          entries={entries}
          onSelect={onSelect}
          disabled={disabled}
          showCount={search.length > 0}
          emptyMessage="No catalog matches. You can keep a custom exercise unclassified."
          renderDetails={(entry) => (
            <>
              <p>Equipment: {entry.equipment}</p>
              <p>Primary: {entry.primaryMuscles.join(", ")}</p>
              <p>
                Secondary: {entry.secondaryMuscles.join(", ") || "None listed"}
              </p>
            </>
          )}
        />
      )}
    </div>
  );
}
