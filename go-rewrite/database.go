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
	dbConnectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbuser, dbpass, dbhost, dbport, dbname)

	fmt.Println("this is the db string")
	fmt.Print(dbConnectionString)
	fmt.Sprintf("this is the dbstring: %s", dbConnectionString)
	fmt.Printf(dbConnectionString)
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		// fmt.Sprintf("this is the dbstring: %s", dbConnectionString)
		panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}
	defer db.Close()
	err = db.Ping()
    if err != nil {
      panic(err.Error()) // proper error handling instead of panic in your app
   }else {
	   fmt.Println("it wrks")
   }
}
