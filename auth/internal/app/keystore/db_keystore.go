package keystore

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sskorolev/balun_microservices/lib/postgres"
)

// DBKeyStore - реализация KeyStore через PostgreSQL (fallback для dev)
type DBKeyStore struct {
	tm        postgres.TransactionManagerAPI
	generator *RSAKeyGenerator
}

func NewDBKeyStore(tm postgres.TransactionManagerAPI) *DBKeyStore {
	return &DBKeyStore{
		tm:        tm,
		generator: NewRSAKeyGenerator(),
	}
}

// GetActiveKey возвращает активный ключ
func (s *DBKeyStore) GetActiveKey(ctx context.Context) (*RSAKey, error) {
	query := `
		SELECT kid, private_key_pem, public_key_pem, status
		FROM rsa_keys
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var key RSAKey
	conn := s.tm.GetQueryEngine(ctx)
	err := conn.QueryRow(ctx, query, KeyStatusActive).Scan(
		&key.KID, &key.PrivateKeyPEM, &key.PublicKeyPEM, &key.Status,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoActiveKey
		}
		return nil, fmt.Errorf("failed to get active key: %w", err)
	}

	return &key, nil
}

// GetKeyByKID возвращает ключ по KID
func (s *DBKeyStore) GetKeyByKID(ctx context.Context, kid string) (*RSAKey, error) {
	query := `
		SELECT kid, private_key_pem, public_key_pem, status
		FROM rsa_keys
		WHERE kid = $1
	`

	var key RSAKey
	conn := s.tm.GetQueryEngine(ctx)
	err := conn.QueryRow(ctx, query, kid).Scan(
		&key.KID, &key.PrivateKeyPEM, &key.PublicKeyPEM, &key.Status,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrKeyNotFound
		}
		return nil, fmt.Errorf("failed to get key by KID: %w", err)
	}

	return &key, nil
}

// GetAllActiveKeys возвращает все активные ключи (active + next)
func (s *DBKeyStore) GetAllActiveKeys(ctx context.Context) ([]*RSAKey, error) {
	query := `
		SELECT kid, private_key_pem, public_key_pem, status
		FROM rsa_keys
		WHERE status IN ($1, $2)
		ORDER BY created_at DESC
	`

	conn := s.tm.GetQueryEngine(ctx)
	rows, err := conn.Query(ctx, query, KeyStatusActive, KeyStatusNext)
	if err != nil {
		return nil, fmt.Errorf("failed to get all active keys: %w", err)
	}
	defer rows.Close()

	var keys []*RSAKey
	for rows.Next() {
		var key RSAKey
		if err := rows.Scan(&key.KID, &key.PrivateKeyPEM, &key.PublicKeyPEM, &key.Status); err != nil {
			return nil, err
		}
		keys = append(keys, &key)
	}

	return keys, nil
}

// CreateKey создает новый RSA ключ
func (s *DBKeyStore) CreateKey(ctx context.Context, status KeyStatus) (*RSAKey, error) {
	// Генерируем RSA ключ
	privateKey, err := s.generator.Generate()
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	privateKeyPEM, err := EncodePrivateKeyToPEM(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encode private key: %w", err)
	}

	publicKeyPEM, err := EncodePublicKeyToPEM(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encode public key: %w", err)
	}

	kid := uuid.New().String()
	now := time.Now()

	query := `
		INSERT INTO rsa_keys (kid, private_key_pem, public_key_pem, status, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING kid, private_key_pem, public_key_pem, status
	`

	key := &RSAKey{}
	conn := s.tm.GetQueryEngine(ctx)
	err = conn.QueryRow(
		ctx, query, kid, privateKeyPEM, publicKeyPEM, status, now,
	).Scan(&key.KID, &key.PrivateKeyPEM, &key.PublicKeyPEM, &key.Status)
	if err != nil {
		return nil, fmt.Errorf("failed to create key: %w", err)
	}

	return key, nil
}

// UpdateKeyStatus обновляет статус ключа
func (s *DBKeyStore) UpdateKeyStatus(ctx context.Context, kid string, status KeyStatus) error {
	query := `
		UPDATE rsa_keys
		SET status = $1
		WHERE kid = $2
	`

	conn := s.tm.GetQueryEngine(ctx)
	_, err := conn.Exec(ctx, query, status, kid)
	if err != nil {
		return fmt.Errorf("failed to update key status: %w", err)
	}

	return nil
}
