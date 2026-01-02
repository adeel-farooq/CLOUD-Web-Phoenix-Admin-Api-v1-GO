## Migrations - Quick Commands (MSSQL)

- Create migration:

  ```
  go run cmd/migrate/main.go create add_orders_table
  ```

  - Creates `db/migrations/NNN_add_orders_table.sql`

- Status:

  ```
  go run cmd/migrate/main.go status
  ```

  - Shows which files are Applied vs Pending (tracked in `schema_migrations`).

- Apply all pending:

  ```
  go run cmd/migrate/main.go up
  ```

  - Applies all `.sql` files under `db/migrations` (recursively), skipping already applied ones.

- Apply specific file:

  ```
  go run cmd/migrate/main.go up 20251229120000_add_users_table.sql
  ```

- Run a single file + record:

  ```
  go run cmd/migrate/main.go run db/migrations/20251229120000_add_users_table.sql
  ```

- Undo a migration (requires .down.sql):
  ```
  go run cmd/migrate/main.go down 20251229120000_add_users_table.sql
  ```

### Notes

- Ensure env vars or `.env` are set: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`.
- CLI uses the [denisenkom/go-mssqldb](https://github.com/denisenkom/go-mssqldb) driver.
- Down migrations convention: create a file with the same name as the migration, but ending with `.down.sql` (e.g., `20251229120000_add_users_table.down.sql`).
