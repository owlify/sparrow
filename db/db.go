package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	_ "github.com/newrelic/go-agent/v3/integrations/nrpgx"
	"github.com/owlify/sparrow/logger"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

const (
	PostgresDriver = "nrpgx"
)

var dbInstance *dbClient

type DBOpts struct {
	URL                   string
	MaxIdleConnection     int
	MaxActiveConnection   int
	MaxConnectionLifetime time.Duration
	DriverName            string
}

type dbClient struct {
	db   *gorm.DB
	opts *DBOpts
}

type DB interface {
	Connect() error
	Close() error
	Get() *gorm.DB
	WithTrx(ctx context.Context, callback trxCallback) (err error)
}

func NewDB(opts *DBOpts) DB {
	dbInstance = &dbClient{
		opts: opts,
	}

	return dbInstance
}

func Get() DB {
	return dbInstance
}

type trxCallback func(ctx context.Context, db DB) error

func (c *dbClient) WithTrx(ctx context.Context, callback trxCallback) (err error) {
	dbTx := c.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			var e error
			if m, ok := r.(string); ok {
				e = errors.New(m)
			} else if m, ok := r.(error); ok {
				e = m
			}
			err = e
			dbTx.Rollback()
		}
	}()

	dbTxClient := &dbClient{
		db: dbTx,
	}

	if err = callback(ctx, dbTxClient); err != nil {
		logger.I(ctx, "Error in runner, rolling back transaction", logger.Field("error", err))
		dbTx.Rollback()
		return err
	}

	if err = dbTx.Commit().Error; err != nil {
		logger.E(ctx, err, "error while committing db transaction")
		return fmt.Errorf("database persistence error: %w", err)
	}

	return nil
}

func (c *dbClient) Connect() error {
	var db *gorm.DB

	gormCfg := &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Info),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: false,
		},
		DisableAutomaticPing: false,
	}

	if c.opts.DriverName == PostgresDriver {
		db = connectPostgres(c.opts, gormCfg)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	if err := sqlDB.Ping(); err != nil {
		return err
	}

	sqlDB.SetMaxIdleConns(c.opts.MaxIdleConnection)
	sqlDB.SetMaxOpenConns(c.opts.MaxActiveConnection)
	sqlDB.SetConnMaxLifetime(c.opts.MaxConnectionLifetime)

	c.db = db

	return nil
}

func (c *dbClient) Close() error {
	sqlDB, err := c.db.DB()
	if err != nil {
		return err
	}
	err = sqlDB.Close()
	if err != nil {
		return err
	}
	return nil
}

func (c *dbClient) Get() *gorm.DB {
	return c.db
}
