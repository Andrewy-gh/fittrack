import { createFileRoute, Link } from '@tanstack/react-router';
import { useState } from 'react';
import { useSuspenseQuery } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { ChevronRight, Plus, Search } from 'lucide-react';
import { Input } from '@/components/ui/input';
import type { ExerciseExerciseResponse } from '@/client';
import { exercisesQueryOptions } from '@/lib/api/exercises';

function ExercisesDisplay({
  exercises,
}: {
  exercises: ExerciseExerciseResponse[];
}) {
  const [searchQuery, setSearchQuery] = useState('');

  const filteredExercises = exercises.filter((exercise) =>
    exercise.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <main>
      <div className="max-w-lg mx-auto space-y-6 px-4 pb-8">
        {/* Header */}
        <div className="flex items-center justify-between pt-4">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Exercises</h1>
          </div>
          <Button size="sm">
            <Plus className="w-4 h-4 mr-2" />
            Add Exercise
          </Button>
        </div>

        {/* MARK: Search */}
        <div className="relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground w-4 h-4" />
          <Input
            placeholder="Search exercises..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-10"
          />
        </div>

        {/* MARK: Exercise List */}
        <Card className="py-0">
          <CardContent className="p-0">
            {filteredExercises.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">
                No exercises found matching "{searchQuery}"
              </div>
            ) : (
              filteredExercises.map((exercise) => (
                <Link
                  to={`/exercises/$exerciseId`}
                  params={{ exerciseId: exercise.id }}
                  key={exercise.id}
                  className="flex items-center justify-between p-4 hover:bg-gray-100/50 transition-colors cursor-pointer border-b border-border last:border-b-0"
                >
                  <h3 className="font-semibold">{exercise.name}</h3>
                  <ChevronRight className="w-5 h-5 text-muted-foreground" />
                </Link>
              ))
            )}
          </CardContent>
        </Card>

        {/* MARK: Results Count */}
        {searchQuery && (
          <div className="text-center text-sm text-muted-foreground">
            {filteredExercises.length} exercise
            {filteredExercises.length !== 1 ? 's' : ''} found
          </div>
        )}
      </div>
    </main>
  );
}

export const Route = createFileRoute('/_auth/exercises/')({
  loader: async ({ context }) => {
    context.queryClient.ensureQueryData(exercisesQueryOptions());
  },
  component: RouteComponent,
});

function RouteComponent() {
  const { data: exercises } = useSuspenseQuery(
    exercisesQueryOptions()
  );  
  return <ExercisesDisplay exercises={exercises} />;
}
