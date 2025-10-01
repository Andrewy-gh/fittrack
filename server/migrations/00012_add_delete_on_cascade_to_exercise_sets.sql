-- +goose Up
-- +goose StatementBegin
ALTER TABLE "set" DROP CONSTRAINT set_exercise_id_fkey;
ALTER TABLE "set" ADD CONSTRAINT set_exercise_id_fkey FOREIGN KEY (exercise_id) REFERENCES exercise(id) ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE "set" DROP CONSTRAINT set_exercise_id_fkey;
ALTER TABLE "set" ADD CONSTRAINT set_exercise_id_fkey FOREIGN KEY (exercise_id) REFERENCES exercise(id);
-- +goose StatementEnd