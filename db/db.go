package db

import (
	"fmt"

	"colly-website/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB(conf *models.Config) error {
	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", conf.MysqlName, conf.MysqlPassword, conf.MysqlPort, conf.DBName)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	DB.AutoMigrate(&models.Task{})
	DB.AutoMigrate(&models.TaskResult{})

	return nil
}
