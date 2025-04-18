package postgresql

import (
	"context"
	"database/sql"
	"time"

	"github.com/1abobik1/token_auth/internal/dto"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type PostgesStorage struct {
	db *sql.DB
}

func NewPostgresStorageProd(storagePath string) (*PostgesStorage, error) {
	db, err := sql.Open("postgres", storagePath)
	if err != nil {
		return nil, err
	}

	return &PostgesStorage{db: db}, nil
}

func NewPostgresForTesting(db *sql.DB) *PostgesStorage {
	return &PostgesStorage{db: db}
}

func (r *PostgesStorage) Close() error {
	return r.db.Close()
}

func (r *PostgesStorage) StoreRefreshToken(ctx context.Context, rec dto.RefreshTokenRecord) error {
	query := `
      INSERT INTO refresh_tokens(user_id, jti, token_hash, client_ip, expires_at)
      VALUES($1,$2,$3,$4,$5)
      ON CONFLICT (user_id, jti) DO UPDATE
        SET token_hash = EXCLUDED.token_hash,
            client_ip = EXCLUDED.client_ip,
            expires_at = EXCLUDED.expires_at;
    `
	
	_, err := r.db.ExecContext(ctx, query,
		rec.UserID, rec.JTI, rec.TokenHash, rec.ClientIP, rec.ExpiresAt,
	)

	return err
}

func (r *PostgesStorage) DeleteRefreshToken(ctx context.Context, userID uuid.UUID, jti string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM refresh_tokens WHERE user_id=$1 AND jti=$2;`,
		userID, jti)

	return err
}

func (r *PostgesStorage) GetRefreshToken(ctx context.Context, userID uuid.UUID, jti string) (dto.RefreshTokenRecord, error) {
	const q = `
    SELECT token_hash, client_ip, expires_at
      FROM refresh_tokens
     WHERE user_id=$1 AND jti=$2 AND expires_at > $3;`

	row := r.db.QueryRowContext(ctx, q, userID, jti, time.Now())

	var rec dto.RefreshTokenRecord
	
	if err := row.Scan(&rec.TokenHash, &rec.ClientIP, &rec.ExpiresAt); err != nil {
		return rec, err
	}

	rec.UserID = userID
	rec.JTI = jti
	
	return rec, nil
}
