package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"fmt"
	"password.exchange/commons"
)
func Connect()  (db *sql.DB) {
	dbhost := commons.GetViperVariable("dbhost")
	dbpass := commons.GetViperVariable("dbpass")
	dbuser := commons.GetViperVariable("dbuser")
	dbname := commons.GetViperVariable("dbname")
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


func Select(uuid string)(msgEncrypted MessagePost){
	db := Connect()
	err := db.QueryRow("select firstname,other_firstname,other_lastname,message,other_email,uniqueid from messages where uniqueid=?", uuid).Scan(&msgEncrypted.FirstName,&msgEncrypted.OtherFirstName,&msgEncrypted.OtherLastName,&msgEncrypted.Content,&msgEncrypted.OtherEmail,&msgEncrypted.Uniqueid)
        
	if err != nil {
		panic(err.Error())
  	}
   return msgEncrypted
}
func Insert(msgEncrypted *Message) {
    db := Connect()

        _, err := db.Exec("INSERT INTO messages(firstname, other_firstname, other_lastname, message, email, other_email, uniqueid) VALUES(?,?,?,?,?,?,?)", msgEncrypted.FirstName,msgEncrypted.OtherFirstName,msgEncrypted.OtherLastName,msgEncrypted.Content,msgEncrypted.Email,msgEncrypted.OtherEmail,msgEncrypted.Uniqueid)
        if err != nil {
            panic(err.Error())
        }
		
}