import { Button } from '@/components/ui/button';
import { Edit, Trash } from 'lucide-react';

export interface ExerciseDetailHeaderProps {
  exerciseName: string;
  onEdit: () => void;
  onDelete: () => void;
}

export function ExerciseDetailHeader({
  exerciseName,
  onEdit,
  onDelete,
}: ExerciseDetailHeaderProps) {
  return (
    <div className="flex items-center justify-between pt-4">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">
          {exerciseName}
        </h1>
      </div>
      <div className="flex flex-col items-center gap-3 md:flex-row">
        <Button size="sm" onClick={onEdit}>
          <Edit className="mr-2 hidden h-4 w-4 md:block" />
          Edit
        </Button>
        <Button
          size="sm"
          variant="outline"
          onClick={onDelete}
          data-testid="delete-exercise-button"
        >
          <Trash className="mr-2 hidden h-4 w-4 md:block" />
          Delete
        </Button>
      </div>
    </div>
  );
}
