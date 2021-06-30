package cservice

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// DatabaseConfig defines the settings required to initialise a MySQL database
// connection.
type DatabaseConfig struct {
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

	// ExtraConfig defines the GORM configuration options.
	ExtraConfig *gorm.Config
}

var db *gorm.DB

func createDSN(config *DatabaseConfig) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", config.User, config.Password, config.Host, config.Port, config.Database)
}

// "root:root@tcp(localhost:3306)/user-service?charset=utf8&parseTime=True&loc=Local", &gorm.Config{}
func InitDatabase(config *DatabaseConfig) error {
	if config.ExtraConfig == nil {
		config.ExtraConfig = &gorm.Config{}
	}

	dsn := createDSN(config)

	var err error
	db, err = gorm.Open(mysql.Open(dsn), config.ExtraConfig)

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
