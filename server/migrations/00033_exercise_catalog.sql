-- +goose Up
-- No name matching or backfill: existing exercises remain unclassified.
ALTER TABLE exercise ADD COLUMN catalog_id TEXT CHECK (catalog_id IN (
    'Barbell_Bench_Press_-_Medium_Grip',
    'Barbell_Curl',
    'Barbell_Deadlift',
    'Barbell_Full_Squat',
    'Bent_Over_Barbell_Row',
    'Crunches',
    'Dumbbell_Bench_Press',
    'Dumbbell_Lunges',
    'Dumbbell_Shoulder_Press',
    'Hammer_Curls',
    'Incline_Dumbbell_Press',
    'Leg_Extensions',
    'Leg_Press',
    'Lying_Leg_Curls',
    'One-Arm_Dumbbell_Row',
    'Plank',
    'Pullups',
    'Pushups',
    'Romanian_Deadlift',
    'Seated_Cable_Rows',
    'Seated_Calf_Raise',
    'Side_Lateral_Raise',
    'Standing_Calf_Raises',
    'Triceps_Pushdown',
    'Wide-Grip_Lat_Pulldown'
));

-- +goose Down
ALTER TABLE exercise DROP COLUMN catalog_id;
