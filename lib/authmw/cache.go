package authmw

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// JWKSCache кеширует JWKS от auth сервиса с автообновлением
type JWKSCache struct {
	mu            sync.RWMutex
	jwks          *JWKS
	jwksURL       string
	refreshPeriod time.Duration
	httpClient    *http.Client
	stopCh        chan struct{}
}

// JWKSCacheConfig конфигурация для JWKSCache
type JWKSCacheConfig struct {
	JWKSURL       string        // URL JWKS endpoint (например, "http://auth:8082/jwks")
	RefreshPeriod time.Duration // Период обновления (по умолчанию 5 минут)
	HTTPTimeout   time.Duration // Таймаут HTTP запросов (по умолчанию 10 секунд)
}

// NewJWKSCache создает новый JWKS кеш с автообновлением
func NewJWKSCache(cfg JWKSCacheConfig) (*JWKSCache, error) {
	if cfg.RefreshPeriod == 0 {
		cfg.RefreshPeriod = 5 * time.Minute
	}
	if cfg.HTTPTimeout == 0 {
		cfg.HTTPTimeout = 10 * time.Second
	}

	cache := &JWKSCache{
		jwksURL:       cfg.JWKSURL,
		refreshPeriod: cfg.RefreshPeriod,
		httpClient: &http.Client{
			Timeout: cfg.HTTPTimeout,
		},
		stopCh: make(chan struct{}),
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
func (c *JWKSCache) GetJWKS() *JWKS {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.jwks
}

// GetKeyByKID возвращает InternalJWK по KID
func (c *JWKSCache) GetKeyByKID(kid string) (*InternalJWK, error) {
	jwks := c.GetJWKS()
	if jwks == nil {
		return nil, fmt.Errorf("JWKS not loaded")
	}
	return jwks.GetKeyByKID(kid)
}

// Stop останавливает автообновление кеша
func (c *JWKSCache) Stop() {
	close(c.stopCh)
}

// refresh обновляет JWKS из auth сервиса
func (c *JWKSCache) refresh(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.jwksURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("JWKS endpoint returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var jwks JWKS
	if err := json.Unmarshal(body, &jwks); err != nil {
		return fmt.Errorf("failed to unmarshal JWKS: %w", err)
	}

	c.mu.Lock()
	c.jwks = &jwks
	c.mu.Unlock()

	return nil
}

// startRefreshLoop запускает фоновый цикл обновления JWKS
func (c *JWKSCache) startRefreshLoop() {
	ticker := time.NewTicker(c.refreshPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), c.httpClient.Timeout)
			if err := c.refresh(ctx); err != nil {
				// В production окружении логируем ошибку
				// Здесь просто игнорируем, используем старый кеш
				_ = err
			}
			cancel()
		case <-c.stopCh:
			return
		}
	}
}
