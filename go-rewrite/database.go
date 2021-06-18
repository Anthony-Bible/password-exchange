package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"fmt"
	"net"
)
func Connect(){
	dbhost := GetViperVariable("dbhost")
	dbpass := GetViperVariable("dbpass")
	dbuser := GetViperVariable("dbuser")
	dbname := GetViperVariable("dbname")
	// dbport := GetViperVariable("dbport")	
	dbConnectionString := fmt.Sprintf("%s:%s@(%s)/%s", dbuser, dbpass, dbhost, dbname)

	fmt.Sprintf("this is the dbstring: %s", dbConnectionString)
	db, err := sql.Open("mysql", dbConnectionString)
	if err != nil {
		panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}
	defer db.Close()
	err = db.Ping()
    if err != nil {
      panic(err.Error()) // proper error handling instead of panic in your app
   }
}
