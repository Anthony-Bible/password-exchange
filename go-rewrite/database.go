package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"fmt"
)
func Connect()  (db *sql.DB) {
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
	return db
}

// func Select(id string	){
// 	dbconn=Connect()
// 	id := 1
	
//     sqlStatement := `SELECT * FROM my_table WHERE id=$1`
//     row := db.QueryRow(sqlStatement, id)
//     err := row.Scan(&col)
//     if err != nil {
//       if err == sql.ErrNoRows {
//           fmt.Println("Zero rows found")
//       } else {
//           panic(err)
//       }
//     }
// }


func Insert(msgEncrypted *Message) {
    db := Connect()

        name := r.FormValue("name")
        city := r.FormValue("city")
        _, err := db.exec("INSERT INTO messages(firstname, lastname, other_firstname, other_lastname, message, email, other_email, uniqueid) VALUES(?,?,?,?,?,?,?,?)", msgEncryptedEncrypted.Firstname,msgEncrypted.Lastname,msgEncrypted.OtherFirstName,msgEncrypted.Content,msgEncrypted.Email,msgEncrypted.OtherEmail,msgEncrypted.Uniqueid)
        if err != nil {
            panic(err.Error())
        }
}