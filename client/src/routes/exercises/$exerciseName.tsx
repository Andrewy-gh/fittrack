import { createFileRoute } from '@tanstack/react-router';
import { fetchExerciseSets } from '../../lib/api/exercises';

export const Route = createFileRoute('/exercises/$exerciseName')({
  component: ExercisePage,
  loader: ({ params }) => fetchExerciseSets(params.exerciseName),
});

function ExercisePage() {
  const sets = Route.useLoaderData();
  const { exerciseName } = Route.useParams();

  return (
    <div className="container mx-auto p-4">
      <h1 className="text-2xl font-bold mb-4">Exercise: {decodeURIComponent(exerciseName)}</h1>
      <div className="space-y-4">
        {sets.map((set) => (
          <div key={set.id} className="p-4 border rounded shadow">
            <p>Weight: {set.weight} lbs</p>
            <p>Reps: {set.reps}</p>
            <p>Type: {set.set_type}</p>
            <p className="text-sm text-gray-500">
              {new Date(set.created_at).toLocaleDateString()}
            </p>
          </div>
        ))}
      </div>
    </div>
  );
}
