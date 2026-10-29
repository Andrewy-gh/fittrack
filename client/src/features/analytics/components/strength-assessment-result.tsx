import { Link } from "@tanstack/react-router";
import type { StrengthAssessment } from "@/features/analytics/api/training-assessment";

/** Present logged strength counts, session findings and reference limitations. */
export function StrengthAssessmentResult({
  assessment,
  timezone,
}: {
  readonly assessment: StrengthAssessment;
  readonly timezone: string;
}) {
  return (
    <>
      <p className="text-sm">
        Reference: <strong>{assessment.reference}</strong>
      </p>
      <p className="text-sm text-muted-foreground">
        {assessment.workingSets} working{" "}
        {assessment.workingSets === 1 ? "set" : "sets"} ·{" "}
        {assessment.trainingDays} training{" "}
        {assessment.trainingDays === 1 ? "day" : "days"}
      </p>
      {assessment.period.partial ? (
        <p className="text-sm">Week in progress · counts so far only.</p>
      ) : assessment.sessions.length === 0 ? (
        <p className="text-sm">
          Cannot assess · no sessions logged for this exercise.
        </p>
      ) : (
        <p className="text-sm">
          {assessment.sessionsMeetingReference} of {assessment.sessions.length}{" "}
          logged sessions reached 2 working sets.
        </p>
      )}
      <ul className="divide-y">
        {assessment.sessions.map((session) => (
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
            {!assessment.period.partial && (
              <p className="text-sm font-medium">
                {session.state === "reference_met"
                  ? "Reference met in logged training"
                  : session.state === "below_reference"
                    ? "Below reference"
                    : "Cannot assess"}
              </p>
            )}
            {!assessment.period.partial && (
              <p className="text-xs text-muted-foreground">{session.reason}</p>
            )}
          </li>
        ))}
      </ul>
      <details className="text-sm">
        <summary className="cursor-pointer font-medium">
          About this comparison
        </summary>
        <div className="space-y-3 pt-3 text-muted-foreground">
          <p>{assessment.applicability}</p>
          {assessment.limitations.map((limitation) => (
            <p key={limitation}>{limitation}</p>
          ))}
          <p>
            Catalog classification is not required for exercise-level set
            counts.
          </p>
          <a
            className="underline underline-offset-4"
            href={assessment.sourceUrl}
            target="_blank"
            rel="noreferrer"
          >
            ACSM reference
          </a>
        </div>
      </details>
    </>
  );
}
