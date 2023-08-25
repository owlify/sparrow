package db

import (
	"time"

	_ "github.com/newrelic/go-agent/v3/integrations/nrpgx"
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
