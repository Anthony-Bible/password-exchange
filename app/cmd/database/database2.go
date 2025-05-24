package database

import (
	"context"
	"database/sql"
	"fmt"
	"net"

	"github.com/Anthony-Bible/password-exchange/app/internal/shared/config"
	storageDomain "github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/domain"
	storageGRPC "github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/adapters/primary/grpc"
	storageMySQL "github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/adapters/secondary/mysql"
	db "github.com/Anthony-Bible/password-exchange/app/pkg/pb/database"

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
	// Use hexagonal architecture
	conf.startHexagonalServer()
}

func (conf Config) startHexagonalServer() {
	address := "0.0.0.0:50051"
	
	// Create database configuration from PassConfig
	dbConfig := storageDomain.DatabaseConfig{
		Host:     conf.PassConfig.DbHost,
		User:     conf.PassConfig.DbUser,
		Password: conf.PassConfig.DbPass,
		Name:     conf.PassConfig.DbName,
	}
	
	// Create MySQL adapter (secondary adapter)
	mysqlAdapter := storageMySQL.NewMySQLAdapter(dbConfig)
	
	// Create storage service (domain)
	storageService := storageDomain.NewStorageService(mysqlAdapter)
	
	// Create gRPC server (primary adapter)
	grpcServer := storageGRPC.NewGRPCServer(storageService, address)
	
	// Start the server
	log.Info().Str("address", address).Msg("Starting storage service with hexagonal architecture")
	if err := grpcServer.Start(); err != nil {
		log.Fatal().Err(err).Msg("Failed to start hexagonal storage server")
	}
}

// Legacy methods kept for backward compatibility
func (conf Config) startLegacyServer() {
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
