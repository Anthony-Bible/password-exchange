package database

import (
	"database/sql"
	"fmt"

	"github.com/Anthony-Bible/password-exchange/app/commons"
	"github.com/Anthony-Bible/password-exchange/app/message"
	"github.com/rs/zerolog/log"

	_ "github.com/go-sql-driver/mysql"
)

func Connect() (db *sql.DB) {
	dbhost, err := commons.GetViperVariable("dbhost")
	if err != nil {
		panic(err)
	}
	dbpass, err := commons.GetViperVariable("dbpass")
	if err != nil {
		panic(err)
	}
	dbuser, err := commons.GetViperVariable("dbuser")
	if err != nil {
		panic(err)
	}
	dbname, err := commons.GetViperVariable("dbname")
	if err != nil {
		panic(err)
	}
	// dbport := commons.GetViperVariable("dbport")
	dbConnectionString := fmt.Sprintf("%s:%s@(%s)/%s?parseTime=true", dbuser, dbpass, dbhost, dbname)

	db, err = sql.Open("mysql", dbConnectionString)
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
func Select(uuid string) (msgEncrypted message.Message) {
	db := Connect()
	err := db.QueryRow("select message,uniqueid from messages where uniqueid=?", uuid).Scan(&msgEncrypted.Content, &msgEncrypted.Uniqueid)

	if err != nil {
		log.Error().Err(err).Msg("Something went wrong with selecting from database")
	}
	return msgEncrypted
}

//Insert encrypted information into database (this is base64 encoded)
func Insert(msgEncrypted *message.Message) {
	db := Connect()

	_, err := db.Exec("INSERT INTO messages( message, uniqueid) VALUES(?,?)", msgEncrypted.Content, msgEncrypted.Uniqueid)
	if err != nil {
		log.Error().Err(err).Msg("Something went wrong with Inserting into database")
	}

}
