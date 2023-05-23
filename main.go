package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	logger     *zap.Logger
	crLoggErr  error
	configPath string
	endPoints  []interface{}
	port       int
)

func RequestHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		w.WriteHeader(http.StatusOK)
	default:
		logger.Info("not supported ", zap.String("method", r.Method), zap.String("uri", r.RequestURI))
		w.WriteHeader(http.StatusBadRequest)
	}

}
func Propagate(w http.ResponseWriter, r *http.Request) {
	logger.Info("Serving request", zap.String("origin", r.Host))
	allResponse := make([]map[string]string, 0)
	for _, val := range endPoints {
		if aMap, correct := val.(map[string]interface{}); correct {
			if aVal, exists := aMap["url"]; exists {
				logger.Info("accessing url ", zap.String("name", aVal.(string)))
				aContent := make(map[string]string)
				aContent["url"] = aVal.(string)

				if resp, errorGet := http.Get(aVal.(string)); errorGet == nil {
					defer resp.Body.Close()
					aContent["status"] = resp.Status
				} else {
					logger.Error("error while performing request", zap.String("url", aContent["url"]), zap.Error(errorGet))
					aContent["status"] = "NOK"
				}

				allResponse = append(allResponse, aContent)
			}
		}

	}
	logger.Info("responding with", zap.Any("all urls", allResponse))
	if contentBytes, errorMars := json.Marshal(allResponse); errorMars != nil {
		logger.Error("error while marhalling", zap.Any("content", allResponse), zap.Error(errorMars))
	} else {
		w.Write([]byte(contentBytes))
	}

}
func init() {

	config := zap.NewDevelopmentConfig()
	config.Level.SetLevel(zap.InfoLevel)
	if logger, crLoggErr = config.Build(); crLoggErr != nil {
		panic("error when setup the log")
	}
	flag.StringVar(&configPath, "config", "config/sam-ping.yaml", "a config string variable")
	flag.Parse()
	viper.SetConfigFile(configPath)
	if errRead := viper.ReadInConfig(); errRead != nil {
		logger.Error("error while loading config", zap.String("config file", configPath), zap.Error(errRead))
	}
	port = viper.GetInt("port")

	anEndpoints := viper.Get("endPoints")
	var ok bool
	if endPoints, ok = anEndpoints.([]interface{}); ok {
		logger.Info("Reading configuration", zap.Int("port", port), zap.Any("endpoints", endPoints))

	}

}
func main() {
	router := mux.NewRouter()
	router.HandleFunc("/ping", RequestHandler)
	router.HandleFunc("/propagate", Propagate)
	address := fmt.Sprintf(":%d", port)
	go func() {
		//start in another go routing
		logger.Info("starting http server", zap.String("address", address))

		if errListen := http.ListenAndServe(address, router); errListen != nil {
			logger.Error("Error starting http server", zap.String("address", address), zap.Error(errListen))
		}

	}()
	//log.Fatal(http.ListenAndServe(address, router))
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch

	logger.Info("Stopping server")
}
