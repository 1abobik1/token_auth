package postgresql_test

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/1abobik1/token_auth/internal/dto"
	"github.com/1abobik1/token_auth/internal/storage/postgresql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestStoreRefreshToken(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	store := postgresql.NewPostgresForTesting(db)

	rec := dto.RefreshTokenRecord{
		UserID:    uuid.New(),
		JTI:       "some-jti",
		TokenHash: "hashed-token",
		ClientIP:  "192.168.1.1",
		ExpiresAt: time.Now().Add(time.Hour),
	}

	mock.ExpectExec(regexp.QuoteMeta(`
      INSERT INTO refresh_tokens(user_id, jti, token_hash, client_ip, expires_at)
      VALUES($1,$2,$3,$4,$5)
      ON CONFLICT (user_id, jti) DO UPDATE
        SET token_hash = EXCLUDED.token_hash,
            client_ip = EXCLUDED.client_ip,
            expires_at = EXCLUDED.expires_at;
    `)).
		WithArgs(rec.UserID, rec.JTI, rec.TokenHash, rec.ClientIP, rec.ExpiresAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = store.StoreRefreshToken(context.Background(), rec)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteRefreshToken(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	store := postgresql.NewPostgresForTesting(db)

	userID := uuid.New()
	jti := "some-jti"

	mock.ExpectExec(`DELETE FROM refresh_tokens WHERE user_id=\$1 AND jti=\$2;`).
		WithArgs(userID, jti).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = store.DeleteRefreshToken(context.Background(), userID, jti)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetRefreshToken(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	store := postgresql.NewPostgresForTesting(db)

	userID := uuid.New()
	jti := "test-jti"
	now := time.Now().Add(30 * time.Minute)

	mock.ExpectQuery(`SELECT token_hash, client_ip, expires_at FROM refresh_tokens`).
		WithArgs(userID, jti, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"token_hash", "client_ip", "expires_at"}).
			AddRow("hash", "127.0.0.1", now))

	rec, err := store.GetRefreshToken(context.Background(), userID, jti)
	require.NoError(t, err)
	require.Equal(t, "hash", rec.TokenHash)
	require.Equal(t, "127.0.0.1", rec.ClientIP)
	require.WithinDuration(t, now, rec.ExpiresAt, time.Second)
	require.Equal(t, userID, rec.UserID)
	require.Equal(t, jti, rec.JTI)

	require.NoError(t, mock.ExpectationsWereMet())
}
