import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { formatExerciseGoalSummary, getExerciseGoal } from '@/lib/exercise-goals';

function ExerciseContextCard({ goalSummary }: { goalSummary: string | null }) {
  if (!goalSummary) {
    return null;
  }

  return (
    <Card className="border-0 shadow-sm backdrop-blur-sm">
      <CardContent className="space-y-4 pt-6">
        <div className="space-y-2">
          <Badge variant="outline" className="border-primary/20 bg-primary/10 text-primary">
            Goal
          </Badge>
          <p className="text-sm font-medium">{goalSummary}</p>
        </div>
      </CardContent>
    </Card>
  );
}

export function ExerciseContextPanel({
  exerciseId,
  exerciseName,
}: {
  exerciseId: number | null;
  exerciseName: string;
}) {
  const goalSummary = formatExerciseGoalSummary(
    getExerciseGoal({ exerciseId, exerciseName })
  );

  if (!exerciseId) {
    return <ExerciseContextCard goalSummary={goalSummary} />;
  }

  return <ExerciseContextCard goalSummary={goalSummary} />;
}
