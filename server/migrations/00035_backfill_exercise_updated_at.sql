-- +goose Up
-- Preserve known edit times. For legacy rows without an edit timestamp, use the
-- only known timestamp (creation), not the time this migration happens to run.
UPDATE exercise SET updated_at = created_at WHERE updated_at IS NULL;
ALTER TABLE exercise ALTER COLUMN updated_at SET NOT NULL;

-- +goose Down
-- Keep populated timestamps: we cannot distinguish backfilled values from real
-- values equal to creation time, so rollback must not erase recorded history.
ALTER TABLE exercise ALTER COLUMN updated_at DROP NOT NULL;
