import { MuscleContributions } from "./muscle-contributions";
import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { addDays, format, parseISO, startOfWeek } from "date-fns";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { trainingAssessmentQueryOptions } from "@/features/analytics/api/training-assessment";
import { trainingProfileQueryOptions } from "@/features/training-profile/api/training-profile";

const coverage = new Map<string, [string, string]>([
  [
    "hypertrophy",
    [
      "Muscle growth",
      "Reviewed muscle contributions are shown. Growth and volume targets are not assessed.",
    ],
  ],
  [
    "endurance",
    ["Endurance", "Muscular and aerobic endurance need different measures."],
  ],
  [
    "general_fitness",
    [
      "General fitness",
      "These logs do not cover all activity needed for this goal.",
    ],
  ],
  [
    "weight_loss",
    ["Weight loss", "Workout logs alone cannot assess weight change."],
  ],
  ["mobility", ["Mobility", "Range of motion is not recorded."]],
]);

export function TrainingAssessment({
  userId,
  exerciseId,
  assessmentWeek,
}: {
  userId: string;
  exerciseId: number;
  assessmentWeek?: string;
}) {
  const profile = useQuery(trainingProfileQueryOptions(userId));
  const goals = profile.data?.goals.length
    ? profile.data.goals
    : profile.data?.primary_goal
      ? [profile.data.primary_goal]
      : [];
  const hasStrength = goals.includes("strength");
  const [weekStart, setWeekStart] = useState(() =>
    assessmentWeek
      ? parseISO(assessmentWeek)
      : addDays(startOfWeek(new Date(), { weekStartsOn: 1 }), -7),
  );
  const currentWeekStart = startOfWeek(new Date(), { weekStartsOn: 1 });
  const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone || "UTC";
  const assessment = useQuery({
    ...trainingAssessmentQueryOptions(
      userId,
      exerciseId,
      format(weekStart, "yyyy-MM-dd"),
      timezone,
    ),
    enabled: hasStrength,
  });

  return (
    <Card id="training-assessment">
      <CardHeader>
        <CardTitle>Training assessment</CardTitle>
      </CardHeader>
      <CardContent className="space-y-5">
        {profile.isPending ? (
          <p
            role="status"
            className="text-sm text-muted-foreground"
          >
            Loading your goals…
          </p>
        ) : profile.isError ? (
          <p
            role="alert"
            className="text-sm"
          >
            Could not load your goals.{" "}
            <Button
              variant="link"
              onClick={() => profile.refetch()}
            >
              Retry
            </Button>
          </p>
        ) : goals.length === 0 ? (
          <p className="text-sm text-muted-foreground">
            Choose goals in your{" "}
            <Link
              to="/settings/training-profile"
              className="underline underline-offset-4"
            >
              profile
            </Link>{" "}
            to see assessment coverage.
          </p>
        ) : null}
        {(hasStrength || goals.includes("hypertrophy")) && (
          <>
            <div className="flex items-center justify-between gap-2">
              <Button
                variant="ghost"
                size="icon"
                aria-label="Previous assessment week"
                onClick={() => setWeekStart(addDays(weekStart, -7))}
              >
                <ChevronLeft />
              </Button>
              <div className="text-center text-sm">
                <p>
                  {format(weekStart, "MMM d")} –{" "}
                  {format(addDays(weekStart, 6), "MMM d, yyyy")}
                </p>
                <p className="text-xs text-muted-foreground">{timezone}</p>
              </div>
              <Button
                variant="ghost"
                size="icon"
                aria-label="Next assessment week"
                disabled={weekStart >= currentWeekStart}
                onClick={() => setWeekStart(addDays(weekStart, 7))}
              >
                <ChevronRight />
              </Button>
            </div>
          </>
        )}
        {hasStrength && (
          <section
            aria-label="Strength reference"
            className="space-y-4"
          >
            <div>
              <h3 className="font-semibold">Strength reference</h3>
              <p className="text-xs text-muted-foreground">
                General adult reference · logged training only
              </p>
            </div>
            {assessment.isPending ? (
              <p
                role="status"
                className="text-sm text-muted-foreground"
              >
                Loading assessment…
              </p>
            ) : assessment.isError ? (
              <p
                role="alert"
                className="text-sm"
              >
                Could not load this assessment.{" "}
                <Button
                  variant="link"
                  onClick={() => assessment.refetch()}
                >
                  Retry
                </Button>
              </p>
            ) : (
              assessment.data && (
                <>
                  <p className="text-sm">
                    Reference: <strong>{assessment.data.reference}</strong>
                  </p>
                  <p className="text-sm text-muted-foreground">
                    {assessment.data.workingSets} working{" "}
                    {assessment.data.workingSets === 1 ? "set" : "sets"} ·{" "}
                    {assessment.data.trainingDays} training{" "}
                    {assessment.data.trainingDays === 1 ? "day" : "days"}
                  </p>
                  {assessment.data.period.partial ? (
                    <p className="text-sm">
                      Week in progress · counts so far only.
                    </p>
                  ) : assessment.data.sessions.length === 0 ? (
                    <p className="text-sm">
                      Cannot assess · no sessions logged for this exercise.
                    </p>
                  ) : (
                    <p className="text-sm">
                      {assessment.data.sessionsMeetingReference} of{" "}
                      {assessment.data.sessions.length} logged sessions reached
                      2 working sets.
                    </p>
                  )}
                  <ul className="divide-y">
                    {assessment.data.sessions.map((session) => (
                      <li
                        key={session.workoutId}
                        className="space-y-1 py-3"
                      >
                        <div className="flex items-center justify-between gap-3">
                          <Link
                            className="text-sm underline underline-offset-4"
                            to="/workouts/$workoutId"
                            params={{ workoutId: session.workoutId }}
                          >
                            {new Intl.DateTimeFormat(undefined, {
                              month: "short",
                              day: "numeric",
                              timeZone: timezone,
                            }).format(new Date(session.date))}
                          </Link>
                          <span className="text-sm">
                            {session.workingSets} working{" "}
                            {session.workingSets === 1 ? "set" : "sets"}
                            {session.invalidWorkingSets > 0
                              ? ` · ${session.invalidWorkingSets} invalid`
                              : ""}
                          </span>
                        </div>
                        {!assessment.data.period.partial && (
                          <p className="text-sm font-medium">
                            {session.state === "reference_met"
                              ? "Reference met in logged training"
                              : session.state === "below_reference"
                                ? "Below reference"
                                : "Cannot assess"}
                          </p>
                        )}
                        {!assessment.data.period.partial && (
                          <p className="text-xs text-muted-foreground">
                            {session.reason}
                          </p>
                        )}
                      </li>
                    ))}
                  </ul>
                  <details className="text-sm">
                    <summary className="cursor-pointer font-medium">
                      About this comparison
                    </summary>
                    <div className="space-y-3 pt-3 text-muted-foreground">
                      <p>{assessment.data.applicability}</p>
                      {assessment.data.limitations.map((limitation) => (
                        <p key={limitation}>{limitation}</p>
                      ))}
                      <p>
                        Catalog classification is not required for
                        exercise-level set counts.
                      </p>
                      <a
                        className="underline underline-offset-4"
                        href={assessment.data.sourceUrl}
                        target="_blank"
                        rel="noreferrer"
                      >
                        ACSM reference
                      </a>
                    </div>
                  </details>
                </>
              )
            )}
          </section>
        )}
        {goals.includes("hypertrophy") && (
          <MuscleContributions
            userId={userId}
            startDate={format(weekStart, "yyyy-MM-dd")}
            timezone={timezone}
          />
        )}
        {goals.some((goal) => goal !== "strength") && (
          <details
            className="text-sm"
            open={!hasStrength}
          >
            <summary className="cursor-pointer font-medium">
              {hasStrength ? "Other goal coverage" : "Goal coverage"}
            </summary>
            <dl className="space-y-3 pt-3">
              {goals
                .filter((goal) => goal !== "strength")
                .map((goal) => (
                  <div key={goal}>
                    <dt className="font-medium">
                      {coverage.get(goal)?.[0] ?? goal} ·{" "}
                      {goal === "hypertrophy"
                        ? "counts available"
                        : "not assessed"}
                    </dt>
                    <dd className="text-muted-foreground">
                      {coverage.get(goal)?.[1] ??
                        "No supported comparison is available yet."}
                    </dd>
                  </div>
                ))}
            </dl>
          </details>
        )}
      </CardContent>
    </Card>
  );
}
