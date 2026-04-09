-- +goose Up
-- +goose StatementBegin
CREATE TABLE exercise_template (
    id SERIAL PRIMARY KEY,
    slug VARCHAR(256) NOT NULL UNIQUE,
    name VARCHAR(256) NOT NULL,
    instructions TEXT,
    category VARCHAR(128),
    equipment VARCHAR(128),
    primary_muscle_group VARCHAR(128),
    secondary_muscle_groups TEXT[] NOT NULL DEFAULT '{}',
    source VARCHAR(64) NOT NULL,
    source_id VARCHAR(256) NOT NULL,
    CONSTRAINT exercise_template_slug_not_empty CHECK (btrim(slug) <> ''),
    CONSTRAINT exercise_template_name_not_empty CHECK (btrim(name) <> ''),
    CONSTRAINT exercise_template_source_not_empty CHECK (btrim(source) <> ''),
    CONSTRAINT exercise_template_source_id_not_empty CHECK (btrim(source_id) <> ''),
    CONSTRAINT exercise_template_source_source_id_key UNIQUE (source, source_id)
);

ALTER TABLE exercise
ADD COLUMN kind VARCHAR(32) NOT NULL DEFAULT 'custom',
ADD COLUMN template_id INTEGER REFERENCES exercise_template(id),
ADD COLUMN instructions TEXT,
ADD COLUMN equipment VARCHAR(128),
ADD COLUMN primary_muscle_group VARCHAR(128),
ADD COLUMN secondary_muscle_groups TEXT[] NOT NULL DEFAULT '{}';

ALTER TABLE exercise
ADD CONSTRAINT exercise_kind_valid CHECK (kind IN ('custom', 'template_based')),
ADD CONSTRAINT exercise_kind_template_state CHECK (
    (kind = 'custom' AND template_id IS NULL)
    OR (kind = 'template_based' AND template_id IS NOT NULL)
);

CREATE INDEX idx_exercise_template_id ON exercise(template_id);

ALTER TABLE exercise_template ENABLE ROW LEVEL SECURITY;

CREATE POLICY exercise_template_select_policy ON exercise_template
    FOR SELECT TO PUBLIC
    USING (true);

GRANT SELECT ON exercise_template TO PUBLIC;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
REVOKE SELECT ON exercise_template FROM PUBLIC;

DROP POLICY IF EXISTS exercise_template_select_policy ON exercise_template;

ALTER TABLE exercise_template DISABLE ROW LEVEL SECURITY;

DROP INDEX IF EXISTS idx_exercise_template_id;

ALTER TABLE exercise
DROP CONSTRAINT IF EXISTS exercise_kind_template_state,
DROP CONSTRAINT IF EXISTS exercise_kind_valid,
DROP COLUMN IF EXISTS secondary_muscle_groups,
DROP COLUMN IF EXISTS primary_muscle_group,
DROP COLUMN IF EXISTS equipment,
DROP COLUMN IF EXISTS instructions,
DROP COLUMN IF EXISTS template_id,
DROP COLUMN IF EXISTS kind;

DROP TABLE IF EXISTS exercise_template;
-- +goose StatementEnd
