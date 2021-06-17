package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"fmt"
)
func Connect(){
	dbhost := GetViperVariable("dbhost")
	dbpass := GetViperVariable("dbpass")
	dbuser := GetViperVariable("dbuser")
	dbname := GetViperVariable("dbname")
	dbport := GetViperVariable("dbport")
	dbstring := dbuser + ":" + dbpass + "@tcp("  + dbhost  + ":" + dbport + ")/" + dbname
	// id:password@tcp(your-amazonaws-uri.com:3306)/dbname
	fmt.Println("this is the db string")
	fmt.Println(dbstring)
	fmt.Sprintf("this is the dbstring: %s", dbstring)
	db, err := sql.Open("mysql", dbstring)
	if err != nil {
		fmt.Sprintf("this is the dbstring: %s", dbstring)
		panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}
	defer db.Close()
	err = db.Ping()
    if err != nil {
      panic(err.Error()) // proper error handling instead of panic in your app
   }
}
