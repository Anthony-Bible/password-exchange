package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"fmt"
)
func Connect(){
	dbHost := GetViperVariable("dbhost")
	dbPass := GetViperVariable("dbpass")
	dbUser := GetViperVariable("dbuser")
	dbName := GetViperVariable("dbname")
	dbPort := GetViperVariable("dbport")
	// dbstring := dbuser + ":" + dbpass + "@tcp("  + dbhost  + ":" + dbport + ")/" + dbname
	dbConnectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPass, dbHost, dbPort, dbName)
	// id:password@tcp(your-amazonaws-uri.com:3306)/dbname
	fmt.Println("this is the db string")
	fmt.Print(dbConnectionString)
	fmt.Sprintf("this is the dbstring: %s", dbConnectionString)
	db, err := sql.Open("mysql", dbConnectionString)
	if err != nil {
		fmt.Sprintf("this is the dbstring: %s", dbConnectionString)
		panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}
	defer db.Close()
	err = db.Ping()
    if err != nil {
      panic(err.Error()) // proper error handling instead of panic in your app
   }
}
