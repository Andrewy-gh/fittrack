package trainingprofile

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Get(ctx context.Context, userID string) (*ProfileResponse, error)
	Upsert(ctx context.Context, userID string, req UpdateProfileRequest) (*ProfileResponse, error)
}

type repository struct {
	logger  *slog.Logger
	queries *db.Queries
	conn    *pgxpool.Pool
}

func NewRepository(logger *slog.Logger, queries *db.Queries, conn *pgxpool.Pool) Repository {
	return &repository{
		logger:  logger,
		queries: queries,
		conn:    conn,
	}
}

func (r *repository) Get(ctx context.Context, userID string) (*ProfileResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	row, err := r.queries.GetUserTrainingProfile(ctx, userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return emptyProfileResponse(), nil
		}
		return nil, fmt.Errorf("get user training profile: %w", err)
	}

	return profileResponseFromRow(row)
}

func (r *repository) Upsert(ctx context.Context, userID string, req UpdateProfileRequest) (*ProfileResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	availableEquipment, err := encodeStringArray(req.AvailableEquipment)
	if err != nil {
		return nil, fmt.Errorf("encode available equipment: %w", err)
	}
	avoidedExercises, err := encodeStringArray(req.AvoidedExercises)
	if err != nil {
		return nil, fmt.Errorf("encode avoided exercises: %w", err)
	}
	movementLimitations, err := encodeOptionalStringArray(req.MovementLimitations)
	if err != nil {
		return nil, fmt.Errorf("encode movement limitations: %w", err)
	}

	row, err := r.queries.UpsertUserTrainingProfileForSettings(ctx, db.UpsertUserTrainingProfileForSettingsParams{
		UserID:                          userID,
		PrimaryGoal:                     optionalText(req.PrimaryGoal),
		ExperienceLevel:                 optionalText(req.ExperienceLevel),
		PreferredSessionDurationMinutes: optionalInt(req.PreferredSessionDurationMinutes),
		UsualTrainingLocation:           optionalText(req.UsualTrainingLocation),
		AvailableEquipment:              availableEquipment,
		AvoidedExercises:                avoidedExercises,
		MovementLimitations:             movementLimitations,
	})
	if err != nil {
		return nil, fmt.Errorf("upsert user training profile: %w", err)
	}

	return profileResponseFromRow(row)
}

func optionalText(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
}

func optionalInt(value *int32) pgtype.Int4 {
	if value == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: *value, Valid: true}
}

func encodeStringArray(values []string) ([]byte, error) {
	if values == nil {
		values = []string{}
	}
	return json.Marshal(values)
}

func encodeOptionalStringArray(values *[]string) ([]byte, error) {
	if values == nil {
		return nil, nil
	}
	return encodeStringArray(*values)
}

var _ Repository = (*repository)(nil)
