package connection

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Mysql struct {
	DB *gorm.DB
}

func NewMysqlConnection()(*Mysql, error){
	err := godotenv.Load()
	if err != nil{
		log.Println(".env file not found")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("DB_DSN")
	}

	if dsn == "" {
		username := os.Getenv("DB_USERNAME")
		password := os.Getenv("DB_PASSWORD")
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		dbName := os.Getenv("DB_NAME")
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", username, password, host, port, dbName)
	} else {
		// Clean up schema prefix if provided by hosting platforms
		dsn = strings.TrimPrefix(dsn, "mysql://")
	}

	db,err := gorm.Open(mysql.Open(dsn),&gorm.Config{})

	if err !=nil{
		return nil, err
	}

	log.Println("MySql database connection established")

	return &Mysql{DB:db,},nil
}

