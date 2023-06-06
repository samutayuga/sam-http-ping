package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"sam-http-ping/cmd"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/ping", cmd.RequestHandler)
	router.HandleFunc("/propagate", cmd.Propagate)
	address := fmt.Sprintf(":%d", cmd.Port)
	//get app name
	appName := os.Getenv("APP_NAME")

	go func() {
		//start in another go routing
		cmd.Logger.Info("starting http server", zap.String("appName", appName), zap.String("address", address))

		if errListen := http.ListenAndServe(address, router); errListen != nil {
			cmd.Logger.Error("Error starting http server", zap.String("address", address), zap.Error(errListen))
		}

	}()
	//log.Fatal(http.ListenAndServe(address, router))
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch

	cmd.Logger.Info("Stopping server")
}
