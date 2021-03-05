package main

import (
	"database/sql"

	"shareme/cleaner"
	"shareme/share"
	"shareme/www"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	var err error

	// init Database
	share.DB, err = sql.Open("mysql", share.Config.MysqlSource)
	if err != nil {
		panic(err)
	}
	err = share.DB.Ping()
	if err != nil {
		panic(err)
	}
	defer share.DB.Close()

	// Services
	cleaner.Main()
	www.Main()
}
