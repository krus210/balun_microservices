Необходимо доработать сервис @auth

- логин/регистрация с хешированием пароля (bcrypt/argon2id) и хранением хеша;
- выдача access (JWT, RSA-подпись) и одноразовых refresh токенов (rotation + anti-reuse);
- JWKS endpoint в auth и локальная валидация JWT в каждом сервисе (без сетевых вызовов на каждый запрос);
 - общая библиотека/мидлварь, подключаемая во все сервисы.


Я добавил методы /Users/sskorolev/GolandProjects/cources/balun_microservices/auth/api/service.proto
в gateway /Users/sskorolev/GolandProjects/cources/balun_microservices/gateway/proto/api/gateway/service.proto
/Users/sskorolev/GolandProjects/cources/balun_microservices/gateway/cmd/main.go

Прото файлы сгенерировал - аовторная генерация не требуется - только если надо будет поменять что-тов в методах 

1. JWT (access):

alg=RS256 (или RS384/RS512), c kid
claims: iss, sub (user_id), aud (массив сервисов/клиентов), iat, exp, nbf?, jti
рекомендуемые TTL: access 5–15 мин; refresh 7–30 дней (настраивается конфигом)

2. Пароли: хеширование и проверка
      bcrypt с cost ≥ 12 (или argon2id с современными параметрами).
      Сравнение — constant-time.
      Политика пароля (минимум длина 8).
3. Refresh-токены
   На Login/Refresh:

Сгенерировать новый refresh; старый пометить used_at или удалить в хранилище.
Сохранять только хеш refresh (например, SHA-256(token)), сам токен — возвращается клиенту.


4. Access-токены: JWT (RSA), key rotation и JWKS
   Подпись приватным ключом (из @lib/secrets), kid → выбирается активный ключ (status='active').
   Ротация ключей: поддержать одновременно ≥2 ключа (active + next), плавный rollover без падений.
   JWKS endpoint: отдаёт все действующие публичные ключи.
   Кэширование JWKS у клиентов @chat @social @users @gateway
5. Создать Библиотека валидации JWT для сервисов (@lib/authmw)
   Сделать общую библиотеку/мидлварь, которую подключают все сервисы (@auth @chat @social @users @gateway):

Функции/возможности:

gRPC server interceptor (+ HTTP middleware для gateway), который:

Извлекает Authorization: Bearer <access_jwt>.
Кэширует JWKS (LRU + TTL), auto-refresh по kid miss.
Валидирует подпись (RSA), iss, aud, exp/nbf, jti?, alg.
Кладёт в context структуру AuthContext{UserID, Scopes?, RawClaims, Token}.
Возвращает UNAUTHENTICATED/PERMISSION_DENIED при проблемах.
Пример конфигурации:

auth:
enabled: true
issuer: "https://auth:8080"
jwks:
url: "http://auth:8080/v1/auth/jwks"
cacheTtl: 5m
refreshTimeout: 2s
required: true # если false — пропускаем без токена (для публичных методов)
Метрики/трейсы: счётчики успешных/ошибочных валидаций.

обязательно добавь go mod и библиотеку в @go.work
в клиенте в go mod 

require (
...
github.com/sskorolev/balun_microservices/lib/authmw v0.0.0
...
)

replace github.com/sskorolev/balun_microservices/lib/authmw => ../lib/authmw

Поддержать allowlist публичных методов (которые не требуют токена).

6. Ограничения и рекомендации
   Хранить только хеши refresh-токенов (и паролей). Никогда не логируйте их значения.
   Access — не хранить в БД.
   JWT-клеймы: проверяйте iss, exp;
   Rate-limit на Login/Refresh (защититься от bruteforce) — можно использовать серверный лимитер из ДЗ-5.
   Пароли — не короче 8 символов;
   Для refresh можно использовать opaque случайные токены (255 бит+) и хранить SHA-256.
   Моежете добавить device_id в refresh для удобной сессийной аналитики и выборочного logout (опционально).
   Ключи RSA держите в файловых секретах;
   Логируйте только факты (id, kid, result), не секреты и не полные токены.


7. Рнекомендации по библиотекам 
jwk "github.com/lestrrat-go/jwx/v2/jwk"
```go
package main

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const issuer = "best.hotel.com"

var ErrInvalidToken = errors.New("invalid token")

func createAccessToken(u user) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		// стандартные JWT claims
		"iss": issuer,                           // кто выдал токен
		"sub": u.Email,                          // кому выдан токен
		"iat": now.Unix(),                       // время создания токена
		"exp": now.Add(15 * time.Minute).Unix(), // время жизни токена
		// наши произвольные claims
		"user_email": u.Email,
		"user_name":  u.Name,
	}

	privateKey, err := getPrivateKey(kid)
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = kid
	return token.SignedString(privateKey)
}

var (
	refreshTokens = make(map[string]struct{})
	mx            sync.RWMutex
)

// функция провайдер ключа для верификации подписи
func keyFunc() jwt.Keyfunc {
	return func(token *jwt.Token) (interface{}, error) {
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New("JWT missing 'kid' header")
		}

		return getPublicKey(kid)
	}
}

func createRefreshToken(u user) (string, error) {
	tokenID := uuid.New().String() // уникальный ID для refresh токена

	now := time.Now()
	claims := jwt.MapClaims{
		// стандартные JWT claims
		"iss": issuer,                             // кто выдал токен
		"sub": u.Email,                            // кому выдан токен
		"iat": now.Unix(),                         // время создания токена
		"exp": now.Add(7 * 24 * time.Hour).Unix(), // время жизни токена
		"jti": tokenID,                            // "JWT ID" — идентификатор токена
		// наши произвольные claims
		"type": "refresh",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = kid
	privateKey, err := getPrivateKey(kid)
	if err != nil {
		return "", err
	}
	signed, err := token.SignedString(privateKey)
	if err != nil {
		return "", err
	}

	// храним токен
	mx.Lock()
	refreshTokens[tokenID] = struct{}{}
	mx.Unlock()

	return signed, nil
}

func verifyRefreshToken(refreshToken string) (user, error) {
	token, err := jwt.Parse(refreshToken, keyFunc(),
		jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Name}),
		jwt.WithIssuer(issuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return user{}, fmt.Errorf("parse token failed: %w", err)
	}

	if !token.Valid {
		return user{}, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["type"] != "refresh" {
		return user{}, ErrInvalidToken
	}

	tokenID, ok := claims["jti"].(string)
	if !ok {
		return user{}, ErrInvalidToken
	}

	email, ok := claims["sub"].(string)
	if !ok {
		return user{}, ErrInvalidToken
	}

	mx.RLock()
	_, exists := refreshTokens[tokenID]
	mx.RUnlock()

	if !exists {
		return user{}, ErrInvalidToken
	}

	mx.Lock()
	delete(refreshTokens, tokenID) // больше нельзя использовать этот refresh
	mx.Unlock()

	idx := slices.IndexFunc(usersDB, func(u user) bool { return strings.EqualFold(email, u.Email) })
	if idx == -1 {
		return user{}, ErrInvalidToken
	}

	return usersDB[idx], nil
}
```

```go
import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
)

// Загрузка приватного ключа из файла
func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil || block.Type != "PRIVATE KEY" {
		return nil, errors.New("невалидный PEM-файл для приватного ключа")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("ключ не RSA")
	}

	return rsaKey, nil
}
```

```go
package main

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const issuer = "best.hotel.com"

var ErrInvalidToken = errors.New("invalid token")

func createAccessToken(u user) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		// стандартные JWT claims
		"iss": issuer,                           // кто выдал токен
		"sub": u.Email,                          // кому выдан токен
		"iat": now.Unix(),                       // время создания токена
		"exp": now.Add(15 * time.Minute).Unix(), // время жизни токена
		// наши произвольные claims
		"user_email": u.Email,
		"user_name":  u.Name,
	}

	privateKey, err := getPrivateKey(kid)
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = kid
	return token.SignedString(privateKey)
}

var (
	refreshTokens = make(map[string]struct{})
	mx            sync.RWMutex
)

// функция провайдер ключа для верификации подписи
func keyFunc() jwt.Keyfunc {
	return func(token *jwt.Token) (interface{}, error) {
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New("JWT missing 'kid' header")
		}

		return getPublicKey(kid)
	}
}

func createRefreshToken(u user) (string, error) {
	tokenID := uuid.New().String() // уникальный ID для refresh токена

	now := time.Now()
	claims := jwt.MapClaims{
		// стандартные JWT claims
		"iss": issuer,                             // кто выдал токен
		"sub": u.Email,                            // кому выдан токен
		"iat": now.Unix(),                         // время создания токена
		"exp": now.Add(7 * 24 * time.Hour).Unix(), // время жизни токена
		"jti": tokenID,                            // "JWT ID" — идентификатор токена
		// наши произвольные claims
		"type": "refresh",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = kid
	privateKey, err := getPrivateKey(kid)
	if err != nil {
		return "", err
	}
	signed, err := token.SignedString(privateKey)
	if err != nil {
		return "", err
	}

	// храним токен
	mx.Lock()
	refreshTokens[tokenID] = struct{}{}
	mx.Unlock()

	return signed, nil
}

func verifyRefreshToken(refreshToken string) (user, error) {
	token, err := jwt.Parse(refreshToken, keyFunc(),
		jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Name}),
		jwt.WithIssuer(issuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return user{}, fmt.Errorf("parse token failed: %w", err)
	}

	if !token.Valid {
		return user{}, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["type"] != "refresh" {
		return user{}, ErrInvalidToken
	}

	tokenID, ok := claims["jti"].(string)
	if !ok {
		return user{}, ErrInvalidToken
	}

	email, ok := claims["sub"].(string)
	if !ok {
		return user{}, ErrInvalidToken
	}

	mx.RLock()
	_, exists := refreshTokens[tokenID]
	mx.RUnlock()

	if !exists {
		return user{}, ErrInvalidToken
	}

	mx.Lock()
	delete(refreshTokens, tokenID) // больше нельзя использовать этот refresh
	mx.Unlock()

	idx := slices.IndexFunc(usersDB, func(u user) bool { return strings.EqualFold(email, u.Email) })
	if idx == -1 {
		return user{}, ErrInvalidToken
	}

	return usersDB[idx], nil
}
```
