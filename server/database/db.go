package database

import (
	"database/sql"
	"fmt"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/driver/sqliteshim"
)

var DB *bun.DB

func InitDB(driver string, connectionString string) *bun.DB {
	if DB == nil {
		switch driver {
		case "sqlite":
			{
				sqldb, err := sql.Open(sqliteshim.ShimName, connectionString)
				if err != nil {
					panic(err)
				}
				DB = bun.NewDB(sqldb, sqlitedialect.New())
				break
			}
		case "postgres":
			{
				sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(connectionString)))
				err := sqldb.Ping()
				if err != nil {
					panic(err)
				}
				DB = bun.NewDB(sqldb, pgdialect.New())
				break
			}
		default:
			panic(fmt.Sprintf("Using %s driver is not implemented yet!", driver))
		}
	}
	return DB
}
