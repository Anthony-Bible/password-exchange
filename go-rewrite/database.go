package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)
func Connect(){
	dbhost := GetViperVariable("dbhost")
	dbpass := GetViperVariable("dbpass")
	dbuser := GetViperVariable("dbuser")
	dbname := GetViperVariable("dbname")
	dbstring := dbuser + ":" + dbpass + "@" + dbhost +"/" +dbname
	db, err := sql.Open("mysql", dbstring)
	if err != nil {
		panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}
	defer db.Close()
	err = db.Ping()
    if err != nil {
      panic(err.Error()) // proper error handling instead of panic in your app
   }
}
