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
	fmt.Println(net.LookupHost(dbhost))
	fmt.Println("this is the db string")
	fmt.Print(dbConnectionString)
	fmt.Sprintf("this is the dbstring: %s", dbConnectionString)
	fmt.Printf(dbConnectionString)
	db, err := sql.Open("mysql", dbConnectionString)
	if err != nil {
		fmt.Printf("this is the dbstring: %s", dbConnectionString)
		panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}
	defer db.Close()
	err = db.Ping()
    if err != nil {
		fmt.Printf("db ping: this is the dbstring: %s", dbConnectionString)
      panic(err.Error()) // proper error handling instead of panic in your app
   }else {
	   fmt.Println("it wrks")
   }
}
