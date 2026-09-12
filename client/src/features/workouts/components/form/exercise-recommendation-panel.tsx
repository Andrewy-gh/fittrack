import { useEffect, useState } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import type { RecommendationResult, RecommendationSnapshot } from "@/client";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent } from "@/components/ui/card";
import {
  recommendationApi,
  type RecommendationApi,
} from "@/features/workouts/api/recommendations";
import type {
  Readiness,
  ExerciseFeedback,
} from "@/features/workouts/utils/recommendation-context";

/** Keeps prescriptions, displayed guidance, and actual set entry independent. */
export function ExerciseRecommendationPanel({
  exerciseId,
  exerciseName,
  userId,
  initialSnapshot,
  onChange,
  api = recommendationApi,
}: {
  exerciseId: number;
  exerciseName: string;
  userId: string;
  initialSnapshot?: RecommendationSnapshot;
  onChange: (
    exerciseName: string,
    recommendation: RecommendationResult | null,
    feedback: ExerciseFeedback,
  ) => void;
  api?: RecommendationApi;
}) {
  const initialReadiness = initialSnapshot?.recommendation.readiness;
  const [readiness, setReadiness] = useState<Readiness>(
    initialReadiness === "sluggish" || initialReadiness === "great"
      ? initialReadiness
      : "normal",
  );
  const initialFeedback = initialSnapshot?.feedback;
  const [feedback, setFeedback] = useState<ExerciseFeedback>(
    initialFeedback === "too_little" ||
      initialFeedback === "about_right" ||
      initialFeedback === "too_much"
      ? initialFeedback
      : "",
  );
  const [minimum, setMinimum] = useState("");
  const [maximum, setMaximum] = useState("");
  const [edited, setEdited] = useState(false);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState("");
  const queryClient = useQueryClient();
  const queryKey = ["exercise-recommendation", userId, exerciseId];
  const query = useQuery({
    queryKey: [...queryKey, readiness],
    queryFn: ({ signal }) => api.get(exerciseId, readiness, signal),
    staleTime: 0,
  });
  const result = query.data?.ok ? query.data.value : null;

  useEffect(() => {
    onChange(exerciseName, result, feedback);
  }, [exerciseName, result, feedback, onChange]);

  useEffect(() => {
    if (!edited && result) {
      const baseline =
        result.source === "prescription" ? result.baseline : null;
      setMinimum(baseline ? String(baseline.min) : "");
      setMaximum(baseline ? String(baseline.max) : "");
    }
  }, [result, edited]);

  async function saveBaseline(clear: boolean) {
    const baseline = clear
      ? null
      : { min: Number(minimum), max: Number(maximum) };
    if (
      baseline &&
      (!Number.isInteger(baseline.min) ||
        !Number.isInteger(baseline.max) ||
        baseline.min < 1 ||
        baseline.max < baseline.min ||
        baseline.max > 20)
    ) {
      setMessage("Enter an ordered range between 1 and 20 working sets.");
      return;
    }
    setSaving(true);
    const outcome = await api.prescribe(exerciseId, baseline);
    if (outcome.ok) {
      await queryClient.invalidateQueries({ queryKey });
      setEdited(false);
      setMessage(
        clear
          ? "Baseline cleared; recent history can be used."
          : "Baseline saved.",
      );
    } else {
      setMessage(outcome.error.message);
    }
    setSaving(false);
  }

  return (
    <Card>
      <CardContent className="space-y-4 pt-6">
        <h2 className="font-semibold">Today’s working sets</h2>
        <p className="text-sm text-muted-foreground">
          Guidance only. Log the sets you actually perform below.
        </p>
        {result?.previous && (
          <p className="text-sm">
            Previous session:{" "}
            {new Date(result.previous.date).toLocaleDateString()} ·{" "}
            {result.previous.workingSets} working sets
          </p>
        )}
        {result && !result.previous && (
          <p className="text-sm text-muted-foreground">
            No previous session for this exercise.
          </p>
        )}
        <fieldset
          className="space-y-2"
          disabled={saving}
        >
          <legend className="text-sm font-medium">
            Optional baseline working-set range
          </legend>
          <p className="text-xs text-muted-foreground">
            Your prescription, or one your coach gave you. Saved to your
            account.
          </p>
          <div className="grid grid-cols-2 gap-3">
            <label className="text-sm">
              Minimum sets
              <Input
                aria-label="Baseline minimum sets"
                type="number"
                min={1}
                max={20}
                value={minimum}
                onChange={(event) => {
                  setEdited(true);
                  setMinimum(event.target.value);
                }}
              />
            </label>
            <label className="text-sm">
              Maximum sets
              <Input
                aria-label="Baseline maximum sets"
                type="number"
                min={1}
                max={20}
                value={maximum}
                onChange={(event) => {
                  setEdited(true);
                  setMaximum(event.target.value);
                }}
              />
            </label>
          </div>
          <div className="flex flex-wrap gap-2">
            <Button
              type="button"
              size="sm"
              onClick={() => void saveBaseline(false)}
            >
              Save baseline
            </Button>
            <Button
              type="button"
              size="sm"
              variant="outline"
              onClick={() => void saveBaseline(true)}
            >
              Clear baseline
            </Button>
          </div>
        </fieldset>
        {message && (
          <p
            role="status"
            className="text-sm"
          >
            {message}
          </p>
        )}
        <fieldset
          className="space-y-2"
          disabled={saving}
        >
          <legend className="text-sm font-medium">Today’s readiness</legend>
          <div className="flex flex-wrap gap-2">
            {(["sluggish", "normal", "great"] as const).map((value) => (
              <label
                key={value}
                className="flex min-h-11 items-center gap-2 rounded-md border px-3 text-sm"
              >
                <input
                  type="radio"
                  name={`readiness-${exerciseId}`}
                  value={value}
                  checked={readiness === value}
                  onChange={() => setReadiness(value)}
                />
                {value === "sluggish"
                  ? "Sluggish"
                  : value === "great"
                    ? "Great"
                    : "Normal"}
              </label>
            ))}
          </div>
        </fieldset>
        <div
          aria-live="polite"
          className="space-y-1 rounded-md bg-muted p-3"
        >
          {query.isPending ? (
            <p>Loading recommendation…</p>
          ) : result ? (
            <>
              <p className="font-semibold">
                {result.range
                  ? `Recommended: ${result.range.min}–${result.range.max} working sets`
                  : "No recommended range yet"}
              </p>
              <p className="text-sm">{result.explanation}</p>
              <p className="text-xs text-muted-foreground">
                Baseline source:{" "}
                {result.source === "prescription"
                  ? "Your saved prescription"
                  : result.source === "history"
                    ? "Recent exercise history"
                    : "None"}
                {result.baseline
                  ? ` (${result.baseline.min}–${result.baseline.max} sets)`
                  : ""}
              </p>
            </>
          ) : (
            <>
              <p>
                Recommendation unavailable. You can still log your actual sets.
              </p>
              <Button
                type="button"
                size="sm"
                variant="outline"
                onClick={() => void query.refetch()}
              >
                Retry guidance
              </Button>
            </>
          )}
        </div>
        <label className="block space-y-1 text-sm">
          After this exercise (optional)
          <select
            className="w-full rounded-md border bg-background p-2"
            value={feedback}
            onChange={(event) => {
              const value = event.target.value;
              if (
                value === "" ||
                value === "too_little" ||
                value === "about_right" ||
                value === "too_much"
              )
                setFeedback(value);
            }}
          >
            <option value="">Not recorded</option>
            <option value="too_little">Too little</option>
            <option value="about_right">About right</option>
            <option value="too_much">Too much</option>
          </select>
        </label>
      </CardContent>
    </Card>
  );
}
