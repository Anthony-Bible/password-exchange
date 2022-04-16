package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"

	db "github.com/Anthony-Bible/password-exchange/app/databasepb"

	"github.com/Anthony-Bible/password-exchange/app/commons"
	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
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

type server struct {
	db.UnimplementedDbServiceServer
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
// func (*server) DecryptMessage(ctx context.Context, request *pb.DecryptedMessageRequest) (*pb.DecryptedMessageResponse, error) {

func (*server) Select(ctx context.Context, request *db.SelectRequest) (*db.SelectResponse, error) {
	dbconnection := Connect()
	response := db.SelectResponse{}
	uuid := request.GetUuid()
	err := dbconnection.QueryRow("select message,uniqueid from messages where uniqueid=?", uuid).Scan(&response.Content, &request.Uuid)

	if err != nil {
		log.Error().Err(err).Msg("Something went wrong with selecting from database")
		return nil, err
	}
	return &response, nil
}

//Insert encrypted information into database (this is base64 encoded)
func (*server) Insert(ctx context.Context, request *db.InsertRequest) (*emptypb.Empty, error) {
	db := Connect()

	_, err := db.Exec("INSERT INTO messages( message, uniqueid) VALUES(?,?)", request.GetContent(), request.GetUuid())
	defer db.Close()
	if err != nil {
		log.Error().Err(err).Msg("Something went wrong with Inserting into database")
	}
	e := &emptypb.Empty{}
	return e, nil

}

func main() {
	// Email:          []byte(ctx.PostForm("email"))
	// FirstName:      []byte(ctx.PostForm("firstname"))
	// OtherFirstName: []byte(ctx.PostForm("other_firstname"))
	// OtherLastName:  []byte(ctx.PostForm("other_lastname"))
	// OtherEmail:     []byte(ctx.PostForm("other_email"))
	// Content:        []byte(ctx.PostForm("content"))

	address := "0.0.0.0:50051"
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal().Err(err).Msg("Problem with starting grpc server")
	}
	// plainMessage := &pb.PlainMessage{
	// Email:          []byte(ctx.PostForm("email")),
	// FirstName:      []byte(ctx.PostForm("firstname")),
	// OtherFirstName: []byte(ctx.PostForm("other_firstname")),
	// OtherLastName:  []byte(ctx.PostForm("other_lastname")),
	// OtherEmail:     []byte(ctx.PostForm("other_email")),
	// Content:        []byte(ctx.PostForm("content"))
	// Url: siteHost + "decrypt/" + msgEncrypted.Uniqueid + "/" + string(encryptionstring[:]),

	// }

	s := grpc.NewServer()
	db.RegisterDbServiceServer(s, &server{})
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatal().Msgf("failed to serve: %v", err)
	}
}
