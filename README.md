### Dependencies
1. Postgres 12
2. Go 1.13
### Configure
1. `cat create.sql | psql < connection string >` (this wipes the specified db and creates tables)
2. `export $(cat .env)`
3. `go run cmd/microjournal/main.go`

