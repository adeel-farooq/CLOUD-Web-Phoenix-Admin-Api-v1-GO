package db

import (
	"database/sql"
	"fmt"
	"os"
	"sync"

	_ "github.com/denisenkom/go-mssqldb"
)

var (
	DB   *sql.DB
	once sync.Once
)

func InitDB() error {
	var err error
	once.Do(func() {
		user := os.Getenv("DB_USER")
		pass := os.Getenv("DB_PASSWORD")
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		name := os.Getenv("DB_NAME")
		if user == "" {
			user = "sa"
		}
		if host == "" {
			host = "localhost"
		}
		if port == "" {
			port = "1433"
		}
		if name == "" {
			name = "master"
		}
		// Print connection string for debugging (without password) after defaults
		safeConnStr := fmt.Sprintf("sqlserver://%s:****@%s:%s?database=%s", user, host, port, name)
		fmt.Println("[DB DEBUG] Connection string:", safeConnStr)
		connStr := fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s", user, pass, host, port, name)
		DB, err = sql.Open("sqlserver", connStr)
		if err == nil {
			err = DB.Ping()
		}
	})
	return err
}
