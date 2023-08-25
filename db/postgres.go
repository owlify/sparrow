package db

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func connectPostgres(opts *DBOpts, gormCfg *gorm.Config) *gorm.DB {

	pgConfig := postgres.Config{
		DriverName: opts.DriverName,
		DSN:        opts.URL,
	}

	db, err := gorm.Open(postgres.New(pgConfig), gormCfg)
	if err != nil {
		panic(err)
	}

	return db
}
