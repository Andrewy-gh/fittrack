-- +goose Up
-- +goose StatementBegin
ALTER TABLE "set" ALTER COLUMN weight TYPE NUMERIC(10,1);
ALTER TABLE "set" ADD CONSTRAINT weight_non_negative CHECK (weight IS NULL OR weight >= 0);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE "set" DROP CONSTRAINT IF EXISTS weight_non_negative;
ALTER TABLE "set" ALTER COLUMN weight TYPE INTEGER USING weight::INTEGER;
-- +goose StatementEnd
