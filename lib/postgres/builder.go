package postgres

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pgxUUID "github.com/vgarvardt/pgx-google-uuid/v5"
)

// Option определяет функциональную опцию для конфигурации postgres
type Option func(*config)

// config содержит настройки для создания Connection и TransactionManager
type config struct {
	// DSN connection string (highest priority if set)
	dsn string

	// Параметры подключения (используются если dsn не указан)
	host     string
	port     int
	database string
	user     string
	password string
	sslMode  string

	// Настройки connection pool
	maxConnIdleTime     time.Duration
	maxConnLifeTime     time.Duration
	minConnectionsCount int32
	maxConnectionsCount int32
	tlsConfig           *tls.Config
}

const (
	maxConnIdleTimeDefault     = time.Minute
	maxConnLifeTimeDefault     = time.Hour
	minConnectionsCountDefault = 2
	maxConnectionsCountDefault = 10
)

// New создает новый Connection pool и TransactionManager с указанными опциями
//
// По умолчанию (без опций подключения) вернет ошибку, так как требуется
// либо DSN строка (WithDSN), либо параметры подключения (WithHost, WithPort, WithDatabase, WithUser, WithPassword)
//
// Примеры:
//
//	// С использованием DSN
//	conn, txMngr, err := postgres.New(ctx,
//	    postgres.WithDSN("postgres://user:password@localhost:5432/mydb?sslmode=disable"),
//	)
//
//	// С использованием отдельных параметров
//	conn, txMngr, err := postgres.New(ctx,
//	    postgres.WithHost("localhost"),
//	    postgres.WithPort(5432),
//	    postgres.WithDatabase("mydb"),
//	    postgres.WithUser("postgres"),
//	    postgres.WithPassword("secret"),
//	    postgres.WithSSLMode("disable"),
//	    postgres.WithMaxConnIdleTime(time.Minute),
//	)
//
//	defer conn.Close()
func New(ctx context.Context, opts ...Option) (*Connection, *TransactionManager, error) {
	cfg := &config{
		maxConnIdleTime:     maxConnIdleTimeDefault,
		maxConnLifeTime:     maxConnLifeTimeDefault,
		minConnectionsCount: minConnectionsCountDefault,
		maxConnectionsCount: maxConnectionsCountDefault,
		sslMode:             "disable",
		port:                5432,
	}

	// Применяем опции
	for _, opt := range opts {
		opt(cfg)
	}

	// Собираем DSN если не указан напрямую
	connString := cfg.dsn
	if connString == "" {
		// Валидация обязательных параметров
		if cfg.host == "" || cfg.database == "" || cfg.user == "" {
			return nil, nil, fmt.Errorf("postgres: missing required connection parameters (host, database, user)")
		}

		// Собираем DSN из параметров
		connString = fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?sslmode=%s",
			cfg.user,
			cfg.password,
			cfg.host,
			cfg.port,
			cfg.database,
			cfg.sslMode,
		)
	}

	// Парсим строку подключения
	connConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, nil, fmt.Errorf("postgres: can't parse connection string: %w", err)
	}

	// Регистрируем UUID тип
	connConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		pgxUUID.Register(conn.TypeMap())
		return nil
	}

	// Применяем настройки pool
	connConfig.MaxConnIdleTime = cfg.maxConnIdleTime
	connConfig.MaxConnLifetime = cfg.maxConnLifeTime
	connConfig.MinConns = cfg.minConnectionsCount
	connConfig.MaxConns = cfg.maxConnectionsCount

	// Применяем TLS конфигурацию если указана
	if cfg.tlsConfig != nil {
		connConfig.ConnConfig.Config.TLSConfig = cfg.tlsConfig
	}

	// Подключаемся к базе данных
	pool, err := pgxpool.NewWithConfig(ctx, connConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("postgres: can't connect to database: %w", err)
	}

	// Проверяем соединение
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, nil, fmt.Errorf("postgres: ping database error: %w", err)
	}

	// Создаем Connection
	conn := &Connection{
		pool: pool,
	}

	// Создаем TransactionManager
	txMngr := NewTransactionManager(conn)

	return conn, txMngr, nil
}

// WithDSN устанавливает полную DSN строку подключения (наивысший приоритет)
//
// Формат DSN:
//
//	postgres://user:password@host:port/database?sslmode=disable&pool_max_conns=10
//
// # Если указана DSN, все остальные параметры подключения (WithHost, WithPort и т.д.) игнорируются
//
// Пример:
//
//	postgres.New(ctx, postgres.WithDSN("postgres://user:pass@localhost:5432/mydb"))
func WithDSN(dsn string) Option {
	return func(c *config) {
		c.dsn = dsn
	}
}

// WithHost устанавливает хост базы данных
//
// Пример:
//
//	postgres.WithHost("localhost")
func WithHost(host string) Option {
	return func(c *config) {
		c.host = host
	}
}

// WithPort устанавливает порт базы данных (по умолчанию: 5432)
//
// Пример:
//
//	postgres.WithPort(5433)
func WithPort(port int) Option {
	return func(c *config) {
		c.port = port
	}
}

// WithDatabase устанавливает имя базы данных
//
// Пример:
//
//	postgres.WithDatabase("mydb")
func WithDatabase(database string) Option {
	return func(c *config) {
		c.database = database
	}
}

// WithUser устанавливает имя пользователя для подключения
//
// Пример:
//
//	postgres.WithUser("postgres")
func WithUser(user string) Option {
	return func(c *config) {
		c.user = user
	}
}

// WithPassword устанавливает пароль для подключения
//
// Пример:
//
//	postgres.WithPassword("secret")
func WithPassword(password string) Option {
	return func(c *config) {
		c.password = password
	}
}

// WithSSLMode устанавливает режим SSL (по умолчанию: "disable")
//
// Возможные значения: disable, require, verify-ca, verify-full
//
// Пример:
//
//	postgres.WithSSLMode("require")
func WithSSLMode(sslMode string) Option {
	return func(c *config) {
		c.sslMode = sslMode
	}
}

// WithMaxConnIdleTime устанавливает максимальное время простоя соединения
// (по умолчанию: 1 минута)
//
// Пример:
//
//	postgres.WithMaxConnIdleTime(5 * time.Minute)
func WithMaxConnIdleTime(d time.Duration) Option {
	return func(c *config) {
		c.maxConnIdleTime = d
	}
}

// WithMaxConnLifeTime устанавливает максимальное время жизни соединения
// (по умолчанию: 1 час)
//
// Пример:
//
//	postgres.WithMaxConnLifeTime(2 * time.Hour)
func WithMaxConnLifeTime(d time.Duration) Option {
	return func(c *config) {
		c.maxConnLifeTime = d
	}
}

// WithMinConnectionsCount устанавливает минимальное количество соединений в пуле
// (по умолчанию: 2)
//
// Пример:
//
//	postgres.WithMinConnectionsCount(5)
func WithMinConnectionsCount(count int32) Option {
	return func(c *config) {
		c.minConnectionsCount = count
	}
}

// WithMaxConnectionsCount устанавливает максимальное количество соединений в пуле
// (по умолчанию: 10)
//
// Пример:
//
//	postgres.WithMaxConnectionsCount(20)
func WithMaxConnectionsCount(count int32) Option {
	return func(c *config) {
		c.maxConnectionsCount = count
	}
}

// WithTLS устанавливает TLS конфигурацию для SSL соединения
//
// Пример:
//
//	tlsConfig := &tls.Config{
//	    InsecureSkipVerify: false,
//	    ServerName:         "postgres.example.com",
//	}
//	postgres.WithTLS(tlsConfig)
func WithTLS(tlsConfig *tls.Config) Option {
	return func(c *config) {
		c.tlsConfig = tlsConfig
	}
}
