package database

import (
	"context"
	"database/sql"
	"fmt"
	"net"

	"github.com/Anthony-Bible/password-exchange/app/config"
	db "github.com/Anthony-Bible/password-exchange/app/databasepb"

	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Config struct {
	db.UnimplementedDbServiceServer
	PassConfig config.PassConfig `mapstructure:",squash"`
}

func (conf *Config) Connect() (db *sql.DB) {
	dbConnectionString := fmt.Sprintf("%s:%s@(%s)/%s?parseTime=true", conf.PassConfig.DbUser, conf.PassConfig.DbPass, conf.PassConfig.DbHost, conf.PassConfig.DbName)

	db, err := sql.Open("mysql", dbConnectionString)
	if err != nil {
		panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}
	return db
}

func (conf *Config) Select(ctx context.Context, request *db.SelectRequest) (*db.SelectResponse, error) {
	dbconnection := conf.Connect()
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
func (conf *Config) Insert(ctx context.Context, request *db.InsertRequest) (*emptypb.Empty, error) {
	db := conf.Connect()

	_, err := db.Exec("INSERT INTO messages( message, uniqueid, other_lastname) VALUES(?,?,?)", request.GetContent(), request.GetUuid(), request.GetPassphrase())
	defer db.Close()
	if err != nil {
		log.Error().Err(err).Msg("Something went wrong with Inserting into database")
	}
	e := &emptypb.Empty{}
	return e, nil

}

func (conf Config) startServer() {

	address := "0.0.0.0:50051"
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal().Err(err).Msg("Problem with starting grpc server")
	}

	s := grpc.NewServer()
	srv := Config{
		PassConfig: conf.PassConfig,
	}
	db.RegisterDbServiceServer(s, &srv)
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatal().Msgf("failed to serve: %v", err)
	}
}
