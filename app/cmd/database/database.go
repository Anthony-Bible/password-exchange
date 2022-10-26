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

func (*server) Select(ctx context.Context, request *db.SelectRequest) (*db.SelectResponse, error) {
	dbconnection := Connect()
	response := db.SelectResponse{}
	uuid := request.GetUuid()
	err := dbconnection.QueryRow("select message,uniqueid,other_lastname from messages where uniqueid=?", uuid).Scan(&response.Content, &request.Uuid, &response.Passphrase)

	if err != nil {
		log.Error().Err(err).Msg("Something went wrong with selecting from database")
		return nil, err
	}
	return &response, nil
}

//Insert encrypted information into database (this is base64 encoded)
func (*server) Insert(ctx context.Context, request *db.InsertRequest) (*emptypb.Empty, error) {
	db := Connect()

	_, err := db.Exec("INSERT INTO messages( message, uniqueid, other_lastname) VALUES(?,?,?)", request.GetContent(), request.GetUuid(), request.GetPassphrase())
	defer db.Close()
	if err != nil {
		log.Error().Err(err).Msg("Something went wrong with Inserting into database")
	}
	e := &emptypb.Empty{}
	return e, nil

}

func main() {

	address := "0.0.0.0:50051"
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal().Err(err).Msg("Problem with starting grpc server")
	}

	s := grpc.NewServer()
	db.RegisterDbServiceServer(s, &server{})
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatal().Msgf("failed to serve: %v", err)
	}
}
