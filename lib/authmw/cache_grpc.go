package authmw

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// AuthServiceClient интерфейс для получения JWKS от auth сервиса
// Возвращает protobuf GetJWKSResponse из jwks.pb.go
type AuthServiceClient interface {
	GetJWKS(ctx context.Context) (*GetJWKSResponse, error)
}

// JWKSCacheGRPC кеширует JWKS от auth сервиса через gRPC с автообновлением
type JWKSCacheGRPC struct {
	mu            sync.RWMutex
	jwks          *JWKS
	client        AuthServiceClient
	refreshPeriod time.Duration
	grpcTimeout   time.Duration
	stopCh        chan struct{}
}

// JWKSCacheGRPCConfig конфигурация для JWKSCacheGRPC
type JWKSCacheGRPCConfig struct {
	Client        AuthServiceClient // Auth service клиент
	RefreshPeriod time.Duration     // Период обновления (по умолчанию 5 минут)
	GRPCTimeout   time.Duration     // Таймаут gRPC запросов (по умолчанию 10 секунд)
}

// NewJWKSCacheGRPC создает новый JWKS кеш с gRPC и автообновлением
func NewJWKSCacheGRPC(cfg JWKSCacheGRPCConfig) (*JWKSCacheGRPC, error) {
	if cfg.RefreshPeriod == 0 {
		cfg.RefreshPeriod = 5 * time.Minute
	}
	if cfg.GRPCTimeout == 0 {
		cfg.GRPCTimeout = 10 * time.Second
	}
	if cfg.Client == nil {
		return nil, fmt.Errorf("auth service client is required")
	}

	cache := &JWKSCacheGRPC{
		client:        cfg.Client,
		refreshPeriod: cfg.RefreshPeriod,
		grpcTimeout:   cfg.GRPCTimeout,
		stopCh:        make(chan struct{}),
	}

	// Первоначальная загрузка JWKS
	if err := cache.refresh(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to fetch initial JWKS: %w", err)
	}

	// Запускаем фоновое обновление
	go cache.startRefreshLoop()

	return cache, nil
}

// GetJWKS возвращает актуальный JWKS
func (c *JWKSCacheGRPC) GetJWKS() *JWKS {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.jwks
}

// GetKeyByKID возвращает InternalJWK по KID
func (c *JWKSCacheGRPC) GetKeyByKID(kid string) (*InternalJWK, error) {
	jwks := c.GetJWKS()
	if jwks == nil {
		return nil, fmt.Errorf("JWKS not loaded")
	}
	return jwks.GetKeyByKID(kid)
}

// Stop останавливает автообновление кеша
func (c *JWKSCacheGRPC) Stop() {
	close(c.stopCh)
}

// refresh обновляет JWKS из auth сервиса через gRPC
func (c *JWKSCacheGRPC) refresh(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, c.grpcTimeout)
	defer cancel()

	// Вызываем GetJWKS через интерфейс
	resp, err := c.client.GetJWKS(ctx)
	if err != nil {
		return fmt.Errorf("failed to invoke GetJWKS: %w", err)
	}

	if len(resp.Jwks) == 0 {
		return fmt.Errorf("received empty JWKS from auth service")
	}

	// Конвертируем protobuf JWK в InternalJWK
	jwks := &JWKS{
		Keys: make([]InternalJWK, 0, len(resp.Jwks)),
	}

	for _, protoJWK := range resp.Jwks {
		jwks.Keys = append(jwks.Keys, InternalJWK{
			KTY: protoJWK.Kty, // protobuf использует маленькие буквы
			Use: protoJWK.Use,
			KID: protoJWK.Kid, // protobuf использует маленькие буквы
			Alg: protoJWK.Alg,
			N:   protoJWK.N,
			E:   protoJWK.E,
		})
	}

	c.mu.Lock()
	c.jwks = jwks
	c.mu.Unlock()

	return nil
}

// startRefreshLoop запускает фоновый цикл обновления JWKS
func (c *JWKSCacheGRPC) startRefreshLoop() {
	ticker := time.NewTicker(c.refreshPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := c.refresh(context.Background()); err != nil {
				// В production окружении логируем ошибку
				// Здесь просто игнорируем, используем старый кеш
				_ = err
			}
		case <-c.stopCh:
			return
		}
	}
}
