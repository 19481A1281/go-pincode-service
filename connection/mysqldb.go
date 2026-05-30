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

// parseDatabaseURL converts standard mysql://user:pass@host:port/db URLs into Go MySQL driver DSN format
func parseDatabaseURL(dbURL string) string {
	if dbURL == "" {
		return ""
	}

	// If it already matches Go MySQL DSN, return as is
	if strings.Contains(dbURL, "@tcp(") {
		return dbURL
	}

	// Strip mysql:// scheme prefix
	cleaned := strings.TrimPrefix(dbURL, "mysql://")

	// Parse user:pass@host:port/dbname
	parts := strings.SplitN(cleaned, "@", 2)
	if len(parts) != 2 {
		return dbURL
	}

	creds := parts[0]
	rest := parts[1]

	dbParts := strings.SplitN(rest, "/", 2)
	if len(dbParts) != 2 {
		return dbURL
	}

	addr := dbParts[0]
	dbAndParams := dbParts[1]

	// Wrap host:port in tcp() as required by Go MySQL driver
	dsn := fmt.Sprintf("%s@tcp(%s)/%s", creds, addr, dbAndParams)

	// Ensure query parameters are present
	if !strings.Contains(dsn, "charset=") {
		if strings.Contains(dsn, "?") {
			dsn += "&charset=utf8mb4&parseTime=True&loc=Local"
		} else {
			dsn += "?charset=utf8mb4&parseTime=True&loc=Local"
		}
	}

	return dsn
}

func NewMysqlConnection()(*Mysql, error){
	err := godotenv.Load()
	if err != nil{
		log.Println(".env file not found")
	}

	dbEnv := os.Getenv("DB_ENV") // "local" or "remote"
	var dsn string

	if dbEnv == "local" {
		username := os.Getenv("DB_USERNAME")
		password := os.Getenv("DB_PASSWORD")
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		dbName := os.Getenv("DB_NAME")
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", username, password, host, port, dbName)
		log.Println("Connecting to LOCAL database...")
	} else {
		// Default to remote URL if DB_ENV is remote or not specified but URL exists
		dsn = os.Getenv("DATABASE_URL")
		if dsn == "" {
			dsn = os.Getenv("DB_DSN")
		}

		if dsn == "" {
			// Fallback to individual variables if no database URL is set at all
			username := os.Getenv("DB_USERNAME")
			password := os.Getenv("DB_PASSWORD")
			host := os.Getenv("DB_HOST")
			port := os.Getenv("DB_PORT")
			dbName := os.Getenv("DB_NAME")
			dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", username, password, host, port, dbName)
			log.Println("Connecting to database using individual credentials...")
		} else {
			dsn = parseDatabaseURL(dsn)
			log.Println("Connecting to REMOTE database via URL...")
		}
	}

	db,err := gorm.Open(mysql.Open(dsn),&gorm.Config{})

	if err !=nil{
		return nil, err
	}

	log.Println("MySql database connection established")

	return &Mysql{DB:db,},nil
}



