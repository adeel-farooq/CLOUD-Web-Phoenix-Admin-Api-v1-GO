package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
)

func loadDotEnv() {
	f, err := os.Open(".env")
	if err != nil {
		return
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			val := strings.Trim(parts[1], `"'`)
			if _, exists := os.LookupEnv(key); !exists {
				os.Setenv(key, val)
			}
		}
	}
}

func mssqlConnString() string {
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
	return fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s", user, pass, host, port, name)
}

func ensureSchemaMigrations(db *sql.DB) error {
	_, err := db.Exec(`IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='schema_migrations' and xtype='U')
	CREATE TABLE schema_migrations (
		id INT IDENTITY(1,1) PRIMARY KEY,
		filename NVARCHAR(255) NOT NULL UNIQUE,
		applied_at DATETIME NOT NULL
	)`)
	return err
}

func appliedMigrations(db *sql.DB) (map[string]bool, error) {
	rows, err := db.Query("SELECT filename FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	applied := make(map[string]bool)
	for rows.Next() {
		var fname string
		rows.Scan(&fname)
		applied[fname] = true
	}
	return applied, nil
}

func listMigrationFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".sql") {
			rel, _ := filepath.Rel(dir, path)
			files = append(files, rel)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

func createMigration(dir, name string) (string, error) {
	timestamp := time.Now().Format("20060102150405")
	fname := fmt.Sprintf("%s_%s.sql", timestamp, name)
	path := filepath.Join(dir, fname)
	return path, ioutil.WriteFile(path, []byte("-- Write your migration here\n"), 0644)
}

func applyMigration(db *sql.DB, dir, fname string) error {
	path := filepath.Join(dir, fname)
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	_, err = db.Exec(string(content))
	if err != nil {
		return err
	}
	_, err = db.Exec("INSERT INTO schema_migrations (filename, applied_at) VALUES (?, GETDATE())", fname)
	return err
}

func undoMigration(db *sql.DB, dir, fname string) error {
	// Convention: create a down migration with same name + .down.sql
	downFile := strings.TrimSuffix(fname, ".sql") + ".down.sql"
	path := filepath.Join(dir, downFile)
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("down migration not found: %s", downFile)
	}
	_, err = db.Exec(string(content))
	if err != nil {
		return err
	}
	_, err = db.Exec("DELETE FROM schema_migrations WHERE filename = ?", fname)
	return err
}

func main() {
	flag.Usage = func() {
		fmt.Print(`Usage:
  create <name>         Create new migration
  status                Show applied/pending migrations
  up [file]             Apply all or specific migration
  down <file>           Undo a migration (requires .down.sql)
  run <file>            Run a migration and record it
`)
	}
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		return
	}
	loadDotEnv()
	db, err := sql.Open("sqlserver", mssqlConnString())
	if err != nil {
		fmt.Println("DB open error:", err)
		return
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		fmt.Println("DB ping error:", err)
		return
	}
	migDir := "db/migrations"
	ensureSchemaMigrations(db)
	switch args[0] {
	case "create":
		if len(args) < 2 {
			fmt.Println("Migration name required")
			return
		}
		path, err := createMigration(migDir, args[1])
		if err != nil {
			fmt.Println("Create error:", err)
		} else {
			fmt.Println("Created:", path)
		}
	case "status":
		applied, _ := appliedMigrations(db)
		files, _ := listMigrationFiles(migDir)
		for _, f := range files {
			if applied[f] {
				fmt.Println("APPLIED  ", f)
			} else {
				fmt.Println("PENDING ", f)
			}
		}
	case "up":
		applied, _ := appliedMigrations(db)
		files, _ := listMigrationFiles(migDir)
		if len(args) == 2 {
			// apply specific file
			if applied[args[1]] {
				fmt.Println("Already applied:", args[1])
				return
			}
			if err := applyMigration(db, migDir, args[1]); err != nil {
				fmt.Println("Apply error:", err)
			} else {
				fmt.Println("Applied:", args[1])
			}
			return
		}
		for _, f := range files {
			if !applied[f] {
				if err := applyMigration(db, migDir, f); err != nil {
					fmt.Println("Apply error:", err)
					break
				}
				fmt.Println("Applied:", f)
			}
		}
	case "down":
		if len(args) < 2 {
			fmt.Println("File required for down")
			return
		}
		if err := undoMigration(db, migDir, args[1]); err != nil {
			fmt.Println("Undo error:", err)
			return
		}
		fmt.Println("Undone:", args[1])
	case "run":
		if len(args) < 2 {
			fmt.Println("File required for run")
			return
		}
		if err := applyMigration(db, migDir, args[1]); err != nil {
			fmt.Println("Run error:", err)
			return
		}
		fmt.Println("Ran and recorded:", args[1])
	default:
		flag.Usage()
	}
}
