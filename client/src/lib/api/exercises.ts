const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api';

export interface ExerciseSet {
  id: number;
  exercise_id: number;
  workout_id: number;
  weight: number;
  reps: number;
  set_type: string;
  created_at: string;
  updated_at: string;
}

export async function fetchExerciseSets(exerciseName: string): Promise<ExerciseSet[]> {
  const response = await fetch(`${API_BASE_URL}/exercises/${encodeURIComponent(exerciseName)}/sets`);
  
  if (!response.ok) {
    throw new Error('Failed to fetch exercise sets');
  }
  
  return response.json();
}
