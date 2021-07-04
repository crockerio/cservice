package cservice

import (
	"errors"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DatabaseConfig defines the settings required to initialise a database
// connection through the GORM ORM.
type DatabaseConfig struct {
	// Driver to use.
	Driver string

	// User to connect to the database as.
	User string

	// Password to authenticate the user.
	Password string

	// Host URL to connect to.
	Host string

	// Port the server is exposed on.
	Port int

	// Database to use.
	Database string

	// Models to auto-migrate.
	Models []interface{}

	// File location to store the database (if we're using the SQLite driver)
	//
	// For a temporary, in-memory database, "file::memory:?cache=shared" can be
	// used.
	File string

	// ExtraConfig defines the GORM configuration options.
	ExtraConfig *gorm.Config
}

var db *gorm.DB

func openMysqlConnection(config *DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", config.User, config.Password, config.Host, config.Port, config.Database)
	return gorm.Open(mysql.Open(dsn), config.ExtraConfig)
}

func openSqliteConnection(config *DatabaseConfig) (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(config.File), config.ExtraConfig)
}

func openConnection(config *DatabaseConfig) (*gorm.DB, error) {
	if config.ExtraConfig == nil {
		config.ExtraConfig = &gorm.Config{}
	}

	switch config.Driver {
	case "mysql":
		return openMysqlConnection(config)
	case "sqlite":
		return openSqliteConnection(config)
	default:
		return nil, errors.New(fmt.Sprintf("unsupported database driver %s", config.Driver))
	}
}

// InitDatabase connection with the given DatabaseConfig
//
// Current supported drivers:
// - MySQL
// - SQLite
func InitDatabase(config *DatabaseConfig) error {
	var err error
	db, err = openConnection(config)

	if err == nil {
		for _, model := range config.Models {
			err := db.AutoMigrate(model)
			if err != nil {
				break
			}
		}
	}

	return err
}

func GetDB() *gorm.DB {
	if db == nil {
		panic("DB not initialised")
	}

	return db
}
