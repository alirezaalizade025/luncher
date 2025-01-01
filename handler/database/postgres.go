package database

import (
	"fmt"
	"luncher/handler/utils"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DbConn struct {
	Conn *gorm.DB
}

var db *DbConn

func Connection() *DbConn {
	if db == nil {
		host := utils.Getenv("DB_HOST", "127.0.0.1")
		port := utils.Getenv("DB_PORT", "5432")
		dbname := utils.Getenv("DB_NAME", "luncher")
		user := utils.Getenv("DB_USER", "postgres")
		password := utils.Getenv("DB_PASSWORD", "postgres")
		sslmode := utils.Getenv("DB_SSLMODE", "disable")

		dsn := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s", host, port, dbname, user, password, sslmode)

		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
			// Logger: logger.Default.LogMode(logger.Silent),
			PrepareStmt: true,
		})
		CheckError(err)

		return &DbConn{Conn: db}
	}
	return db

}


func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}
