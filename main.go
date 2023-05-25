package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type SamPropagator interface {
	doGet(url string) *SamResponse
}
type SamPayload struct {
	httClient *http.Client
}
type SamResponse struct {
	ResponseCode    int
	ResponseMessage string
}

//implement the interface Propagator on Payload type
func (p *SamPayload) doGet(url string) *SamResponse {
	if httpResp, errGet := p.httClient.Get(url); errGet == nil {
		defer httpResp.Body.Close()
		r := SamResponse{ResponseCode: httpResp.StatusCode, ResponseMessage: "OK"}
		logger.Info("Response with", zap.Int("response", r.ResponseCode))
		return &r
	} else {
		logger.Error("error while performing request", zap.String("url", url), zap.Error(errGet))
		return &SamResponse{ResponseCode: -1, ResponseMessage: errGet.Error()}
	}

}

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

func doPropagate(singleMap interface{}, payload *SamPayload, resp chan *SamResponse) {
	if aMap, correct := singleMap.(map[string]interface{}); correct {
		if aVal, exists := aMap["url"]; exists {

			aResp := payload.doGet(aVal.(string))
			logger.Info("response from get url ", zap.String("url", aVal.(string)),
				zap.Int("responseCode", aResp.ResponseCode),
				zap.String("responseMessage", aResp.ResponseMessage))

			resp <- aResp

		}
	}
	//return nil
}
func managePropagation(requestor *SamPayload, allResponseChan chan *SamResponse) {
	for _, _val := range endPoints {
		go func(myVal interface{}) {
			doPropagate(myVal, requestor, allResponseChan)
		}(_val)

	}
}
func Propagate(w http.ResponseWriter, r *http.Request) {
	logger.Info("Serving request", zap.String("origin", r.Host))
	p := SamPayload{httClient: &http.Client{}}

	allResponse := make([]*SamResponse, 0)
	allResponseChan := make(chan *SamResponse, len(endPoints))
	managePropagation(&p, allResponseChan)
	//monitor the response
	for {
		allResponse = append(allResponse, <-allResponseChan)
		logger.Info("receive result", zap.Int("so far received ", len(allResponse)))
		if len(allResponse) >= len(endPoints) {
			logger.Info("completed ...")

			break
		}

		logger.Info("waiting ...")
		time.Sleep(time.Duration(1) * time.Second)
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
