import { useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { Target } from "lucide-react";
import {
  getExercisesByIdGoalCheck,
  type ExerciseStrengthGoalCheck,
} from "@/client";
import { getExercisesByIdGoalCheckQueryOptions } from "@/client/@tanstack/react-query.gen";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";

const number = new Intl.NumberFormat(undefined, { maximumFractionDigits: 1 });

export function ExerciseGoalCheck({
  exerciseId,
  isDemoMode,
  userId,
}: {
  exerciseId: number;
  isDemoMode: boolean;
  userId: string;
}) {
  return (
    <Card aria-labelledby="goal-check-title">
      <CardContent className="space-y-4 pt-6">
        <div className="flex items-center gap-2">
          <Target
            className="h-4 w-4 text-primary"
            aria-hidden="true"
          />
          <h2
            id="goal-check-title"
            className="font-semibold"
          >
            Strength Goal Check
          </h2>
        </div>
        {isDemoMode ? (
          <p className="text-sm text-muted-foreground">
            Sign in and choose strength in your profile to see your check.
          </p>
        ) : (
          <GoalCheckQuery
            exerciseId={exerciseId}
            userId={userId}
          />
        )}
      </CardContent>
    </Card>
  );
}

function GoalCheckQuery({
  exerciseId,
  userId,
}: {
  exerciseId: number;
  userId: string;
}) {
  const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone || "UTC";
  const options = getExercisesByIdGoalCheckQueryOptions({
    path: { id: exerciseId },
    query: { timezone },
  });
  const { data, isPending, isError, refetch } = useQuery({
    queryKey: [...options.queryKey, userId],
    queryFn: ({ signal }) =>
      getExercisesByIdGoalCheck({
        path: { id: exerciseId },
        query: { timezone },
        signal,
        throwOnError: true,
      }).then((response) => response.data),
    staleTime: 0,
    refetchInterval: 60_000,
  });
  if (isPending)
    return (
      <p
        role="status"
        className="text-sm text-muted-foreground"
      >
        Loading Goal Check…
      </p>
    );
  if (isError && !data)
    return (
      <div className="space-y-2">
        <p
          role="alert"
          className="text-sm"
        >
          Couldn’t load this check.
        </p>
        <Button
          variant="outline"
          size="sm"
          onClick={() => void refetch()}
        >
          Retry Goal Check
        </Button>
      </div>
    );
  if (!data) return null;
  return (
    <>
      {isError && (
        <p
          role="status"
          className="text-sm text-muted-foreground"
        >
          Couldn’t refresh. These numbers may be out of date.
        </p>
      )}
      <GoalCheckResult result={data} />
    </>
  );
}

function Comparison({ status }: { status: string }) {
  const met = status === "reference_met";
  return (
    <span
      className={`rounded-full px-2 py-1 text-xs font-medium ${met ? "bg-primary/10 text-primary" : "bg-muted text-muted-foreground"}`}
    >
      {met
        ? "Meets guide"
        : status === "below_reference"
          ? "Below guide"
          : "Not enough data"}
    </span>
  );
}

export function GoalCheckResult({
  result,
}: {
  result: ExerciseStrengthGoalCheck;
}) {
  if (!result.applicable)
    return (
      <div className="space-y-2 text-sm text-muted-foreground">
        <p>Choose strength in your profile to see your check.</p>
        <Link
          to="/settings/training-profile"
          className="text-primary underline underline-offset-4"
        >
          Training profile
        </Link>
      </div>
    );
  const { frequency, working_sets: sets, loading, sessions } = result;
  const date = (value: string) =>
    new Date(value).toLocaleDateString(undefined, {
      month: "short",
      day: "numeric",
      timeZone: result.window.timezone,
    });
  return (
    <>
      <p className="text-xs text-muted-foreground">
        Last 7 days · {date(result.window.start)}–{date(result.window.end)}
      </p>
      <dl className="divide-y">
        <div className="space-y-2 pb-4">
          <div className="flex flex-wrap items-center justify-between gap-2">
            <dt className="text-sm font-medium">Training days</dt>
            <Comparison status={frequency.status} />
          </div>
          <dd>
            <span className="text-2xl font-semibold">
              {frequency.training_days}
            </span>
            <span className="ml-2 text-sm text-muted-foreground">
              across {sessions.length}{" "}
              {sessions.length === 1 ? "session" : "sessions"}
            </span>
          </dd>
          <p className="text-xs text-muted-foreground">
            Guide: {frequency.reference_min}+ days per week
          </p>
        </div>
        <div className="space-y-2 py-4">
          <div className="flex flex-wrap items-center justify-between gap-2">
            <dt className="text-sm font-medium">Working sets per session</dt>
            <Comparison status={sets.status} />
          </div>
          <dd>
            <span className="text-2xl font-semibold">
              {sets.average == null ? "—" : number.format(sets.average)}
            </span>
            <span className="ml-2 text-sm text-muted-foreground">
              average · {sets.total} total
            </span>
          </dd>
          <p className="text-xs text-muted-foreground">
            Guide: {sets.reference_min}–{sets.reference_max} sets · Warm-ups
            excluded
          </p>
        </div>
        <div className="space-y-2 pt-4">
          <dt className="text-sm font-medium">
            Weight{" "}
            <span className="font-normal text-muted-foreground">
              · rough comparison
            </span>
          </dt>
          <dd className="text-sm">
            {loading.reps_min == null
              ? "No working reps logged."
              : `${loading.reps_min}–${loading.reps_max} reps per set`}
          </dd>
          {loading.reference &&
          loading.percent_min != null &&
          loading.percent_max != null ? (
            <>
              <p className="text-sm">
                {number.format(loading.percent_min)}–
                {number.format(loading.percent_max)}% of saved max
              </p>
              <p className="text-xs text-muted-foreground">
                Saved max: {number.format(loading.reference.value)} lb ·{" "}
                {loading.reference.origin === "workout_estimate"
                  ? "estimated"
                  : "entered by you"}
              </p>
            </>
          ) : (
            <p className="text-xs text-muted-foreground">
              Not enough weight data to compare.
            </p>
          )}
        </div>
      </dl>
      {sessions.length < 2 && (
        <p className="rounded-md bg-muted p-3 text-sm">
          {sessions.length === 0
            ? "No working sets logged in the last 7 days."
            : "One session logged. Too soon to see a pattern."}
        </p>
      )}
      {sessions.length > 0 && (
        <details className="text-sm">
          <summary className="cursor-pointer font-medium">
            Your sessions
          </summary>
          <ul className="mt-2 space-y-2">
            {sessions.map((session) => (
              <li
                key={session.workout_id}
                className="flex justify-between gap-3"
              >
                <Link
                  to="/workouts/$workoutId"
                  params={{ workoutId: session.workout_id }}
                  className="text-primary underline underline-offset-4"
                >
                  {date(session.date)}
                </Link>
                <span>
                  {session.working_sets} working{" "}
                  {session.working_sets === 1 ? "set" : "sets"}
                </span>
              </li>
            ))}
          </ul>
        </details>
      )}
      <details className="border-t pt-3 text-sm">
        <summary className="cursor-pointer font-medium">
          About this check
        </summary>
        <div className="mt-3 space-y-3 text-xs text-muted-foreground">
          <p>
            These are guides, not rules. Light weeks and short sessions can be
            intentional. More than three sets isn’t automatically too much.
          </p>
          <p>
            We count this exercise’s working sets over the last 7 days,
            including today so far. Other exercise names aren’t combined. Dates
            use {result.window.timezone}.
          </p>
          <p>
            Saved maxes can be estimated or old. Bodyweight, assistance and
            different machines may not compare. This doesn’t tell us whether
            your weights were heavy enough.
          </p>
          {loading.reference && (
            <p>
              Saved max{" "}
              {loading.reference.updated_at
                ? `updated ${date(loading.reference.updated_at)}`
                : "date unknown"}
              . Compared {loading.sets_with_load} of {sets.total} sets with
              recorded weights.
            </p>
          )}
          <p>Research behind the guides:</p>
          <ul className="space-y-2">
            <li>
              <a
                href="https://pubmed.ncbi.nlm.nih.gov/41843416/"
                target="_blank"
                rel="noreferrer"
                className="underline"
              >
                ACSM position stand (2026)
              </a>
            </li>
            <li>
              <a
                href="https://pubmed.ncbi.nlm.nih.gov/29470825/"
                target="_blank"
                rel="noreferrer"
                className="underline"
              >
                Grgic et al. (2018)
              </a>
            </li>
            <li>
              <a
                href="https://pubmed.ncbi.nlm.nih.gov/41343037/"
                target="_blank"
                rel="noreferrer"
                className="underline"
              >
                Pelland et al. (2026)
              </a>
            </li>
          </ul>
        </div>
      </details>
    </>
  );
}
