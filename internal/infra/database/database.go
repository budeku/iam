package database

import (
	"budeku/iam/internal/conf"
	"budeku/iam/internal/infra/ent/db"
	"budeku/iam/internal/infra/ent/db/migrate"
	_ "budeku/iam/internal/infra/ent/db/runtime"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/go-sql-driver/mysql"
)

const (
	pingTimeout  = 5 * time.Second
	closeTimeout = 5 * time.Second
)

var (
	mu       sync.Mutex
	client   *db.Client
	sqlDB    *sql.DB
	initOnce sync.Once
	initErr  error
)

func Init(cfg conf.Database) error {
	initOnce.Do(func() {
		dsn := cfg.DSN()

		drv, err := entsql.Open(dialect.MySQL, dsn)
		if err != nil {
			initErr = err
			return
		}

		dbConn := drv.DB()
		dbConn.SetMaxIdleConns(cfg.MaxIdleConns)
		dbConn.SetMaxOpenConns(cfg.MaxOpenConns)
		dbConn.SetConnMaxLifetime(cfg.MaxLifetime)
		dbConn.SetConnMaxIdleTime(cfg.MaxIdleTime)

		ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
		defer cancel()
		if err := dbConn.PingContext(ctx); err != nil {
			_ = drv.Close()
			initErr = err
			return
		}

		newClient := db.NewClient(db.Driver(drv))

		if conf.Debug {
			newClient = newClient.Debug()
		}

		mu.Lock()
		client = newClient
		sqlDB = dbConn
		mu.Unlock()
	})

	return initErr
}

func DB() *db.Client {
	mu.Lock()
	defer mu.Unlock()
	return client
}

func Migrate(ctx context.Context) error {
	c := DB()
	if c == nil {
		return fmt.Errorf("数据库连接未初始化")
	}
	if err := c.Schema.Create(
		ctx,
		migrate.WithForeignKeys(false),
	); err != nil {
		return err
	}
	return nil
}

func Close() error {
	mu.Lock()
	c := client
	client = nil
	sqlDB = nil
	mu.Unlock()

	if c == nil {
		return nil
	}

	done := make(chan error, 1)
	go func() {
		done <- c.Close()
	}()

	select {
	case err := <-done:
		if err != nil {
			return err
		}
		return nil
	case <-time.After(closeTimeout):
		return errors.New("关闭连接超时")
	}
}
