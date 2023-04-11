// Package web
package web

import (
	"context"
	"net"
	"net/http"

	"github.com/rs/zerolog/log"
	gcache "github.com/vimeo/galaxycache"
	ghttp "github.com/vimeo/galaxycache/http"
)

//query kubernetes headless service for endpoints
// Uses multiple A records to  get the endpoints
func getEndpoints(svcName string) []string {
	//query kubernetes headless service for endpoints
	endpoints, err := net.LookupIP(svcName)
	if err != nil {
		log.Error().Err(err).Msgf("Something went wrong with looking up %s service", svcName)
	}

	var endpointsString []string
	for _, endpoint := range endpoints {
		endpointsString = append(endpointsString, endpoint.String())
	}
	return endpointsString

}

//Modify string slice to modify each string
func modifyStringSlice(slice []string, prefix, suffix string) []string {
	for i, s := range slice {
		slice[i] = prefix + s + suffix
	}
	return slice
}

//Initialize galaxy cache
func (conf *Config) initGalaxyCache() {
	endpoints := getEndpoints("password-exchange")
	modifiedEndpoints := modifyStringSlice(endpoints, "http://", ":8081")
	httpProto := ghttp.NewHTTPFetchProtocol(nil)
	universe := gcache.NewUniverse(httpProto, endpoints[0])
	//set peers of universe
	universe.Set(modifiedEndpoints...)
	getter := gcache.GetterFunc(func(ctx context.Context, key string, dest gcache.Codec) error {
		uploadID := conf.initiateS3MultipartUpload(key)
		return dest.UnmarshalBinary([]byte(uploadID))
	})
	//Create a new galaxy
	galaxy := universe.NewGalaxy("password-exchange", 1<<20, getter)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serveMux := http.NewServeMux()
	ghttp.RegisterHTTPHandler(universe, nil, serveMux)
	//store galaxy in config struct so I can just call conf.g.get()
	conf.Galaxy = galaxy
	var srv http.Server
	go func() {
		log.Info().Msg("Starting HTTP server on :8081")
		httpAltListener, err := net.Listen("tcp", ":8081")
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to start HTTP server")
		}
		srv.Handler = serveMux
		if err := srv.Serve(httpAltListener); err != nil {
			log.Error().Err(err).Msg("Something went wrong with starting http server")
		}
	}()
	<-ctx.Done()
	srv.Shutdown(ctx)

}
