package helpers

import "github.com/jackc/pgx/v5/pgtype"

func ToPgText(s string) pgtype.Text {
	return pgtype.Text{
		String: s,
		Valid:  true,
	}
}
