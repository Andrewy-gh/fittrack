-- Setup test data for privilege escalation testing

\echo 'Setting up test data for RLS privilege escalation testing...'

-- Insert test users (these would normally be created by the application)
INSERT INTO users (user_id) VALUES ('test-user-1') ON CONFLICT (user_id) DO NOTHING;
INSERT INTO users (user_id) VALUES ('test-user-2') ON CONFLICT (user_id) DO NOTHING;
INSERT INTO users (user_id) VALUES ('test-user-3') ON CONFLICT (user_id) DO NOTHING;

-- Create test workouts for different users
-- User 1 workouts
SET app.current_user_id = 'test-user-1';
INSERT INTO workout (date, notes, user_id) VALUES 
    ('2024-01-01 10:00:00+00', 'User 1 workout A', 'test-user-1'),
    ('2024-01-02 10:00:00+00', 'User 1 workout B', 'test-user-1')
    ON CONFLICT DO NOTHING;

-- User 2 workouts  
SET app.current_user_id = 'test-user-2';
INSERT INTO workout (date, notes, user_id) VALUES 
    ('2024-01-01 11:00:00+00', 'User 2 workout A', 'test-user-2'),
    ('2024-01-02 11:00:00+00', 'User 2 workout B', 'test-user-2')
    ON CONFLICT DO NOTHING;

-- User 3 workouts
SET app.current_user_id = 'test-user-3'; 
INSERT INTO workout (date, notes, user_id) VALUES 
    ('2024-01-01 12:00:00+00', 'User 3 workout A', 'test-user-3'),
    ('2024-01-02 12:00:00+00', 'User 3 workout B', 'test-user-3')
    ON CONFLICT DO NOTHING;

-- Create test exercises for different users
SET app.current_user_id = 'test-user-1';
INSERT INTO exercise (name, user_id) VALUES 
    ('User 1 Push ups', 'test-user-1'),
    ('User 1 Pull ups', 'test-user-1')
    ON CONFLICT (user_id, name) DO NOTHING;

SET app.current_user_id = 'test-user-2';
INSERT INTO exercise (name, user_id) VALUES 
    ('User 2 Squats', 'test-user-2'),
    ('User 2 Bench Press', 'test-user-2')
    ON CONFLICT (user_id, name) DO NOTHING;

SET app.current_user_id = 'test-user-3';
INSERT INTO exercise (name, user_id) VALUES 
    ('User 3 Deadlifts', 'test-user-3'),
    ('User 3 Curls', 'test-user-3')
    ON CONFLICT (user_id, name) DO NOTHING;

\echo 'Test data setup complete!'
\echo 'Created 3 test users with 2 workouts and 2 exercises each.'
