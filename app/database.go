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
	dbConnectionString := fmt.Sprintf("%s:%s@(%s)/%s?parseTime=true", dbuser, dbpass, dbhost, dbname)

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

//Select Get the information based on the uuid from the url
func Select(uuid string)(msgEncrypted Message){
	db := Connect()
	err := db.QueryRow("select firstname,other_firstname,other_lastname,message,email,other_email,uniqueid from messages where uniqueid=?", uuid).Scan(&msgEncrypted.FirstName,&msgEncrypted.OtherFirstName,&msgEncrypted.OtherLastName,&msgEncrypted.Content,&msgEncrypted.Email,&msgEncrypted.OtherEmail,&msgEncrypted.Uniqueid)
        
	if err != nil {
		panic(err.Error())
  	}
   return msgEncrypted
}
//Insert encrypted information into database (this is base64 encoded)
func Insert(msgEncrypted *Message) {
    db := Connect()

        _, err := db.Exec("INSERT INTO messages(firstname, other_firstname, other_lastname, message, email, other_email, uniqueid) VALUES(?,?,?,?,?,?,?)", msgEncrypted.FirstName,msgEncrypted.OtherFirstName,msgEncrypted.OtherLastName,msgEncrypted.Content,msgEncrypted.Email,msgEncrypted.OtherEmail,msgEncrypted.Uniqueid)
        if err != nil {
            panic(err.Error())
        }
		
}