package cservice

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

func InitDatabase(dsn string, config *gorm.Config) error {
	var err error
	db, err = gorm.Open(mysql.Open(dsn), config)
	return err
}

func MigrateModels(models ...interface{}) {
	for _, model := range models {
		err := db.AutoMigrate(model)
		if err != nil {
			panic(err)
		}
	}
}
