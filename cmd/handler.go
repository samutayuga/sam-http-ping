package cmd

import (
	"encoding/json"
	"flag"
	"net/http"
	"os"
	"time"

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
	Origin          string
	Destination     string
}

//implement the interface Propagator on Payload type
func (p *SamPayload) doGet(url string) *SamResponse {
	var hostN string
	var errH error
	if hostN, errH = os.Hostname(); errH != nil {
		Logger.Error("cannot get hostname", zap.Error(errH))
	}
	if httpResp, errGet := p.httClient.Get(url); errGet == nil {
		defer httpResp.Body.Close()

		r := SamResponse{ResponseCode: httpResp.StatusCode, ResponseMessage: "OK", Origin: hostN, Destination: url}
		Logger.Info("Response with", zap.Int("response", r.ResponseCode))
		return &r
	} else {
		Logger.Error("error while performing request", zap.String("url", url), zap.Error(errGet))
		return &SamResponse{ResponseCode: -1, ResponseMessage: errGet.Error(), Origin: hostN, Destination: url}
	}

}

var (
	Logger     *zap.Logger
	crLoggErr  error
	configPath string
	endPoints  []interface{}
	Port       int
)

func RequestHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		w.WriteHeader(http.StatusOK)
	default:
		Logger.Info("not supported ", zap.String("method", r.Method), zap.String("uri", r.RequestURI))
		w.WriteHeader(http.StatusBadRequest)
	}

}

func doPropagate(singleMap interface{}, payload *SamPayload, resp chan *SamResponse) {
	if aMap, correct := singleMap.(map[string]interface{}); correct {
		if aVal, exists := aMap["url"]; exists {

			aResp := payload.doGet(aVal.(string))
			Logger.Info("response from get url ", zap.String("url", aVal.(string)),
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
	Logger.Info("Serving request", zap.String("origin", r.Host))
	p := SamPayload{httClient: &http.Client{}}

	allResponse := make([]*SamResponse, 0)
	allResponseChan := make(chan *SamResponse, len(endPoints))
	managePropagation(&p, allResponseChan)
	//monitor the response
	for {
		allResponse = append(allResponse, <-allResponseChan)
		Logger.Info("receive result", zap.Int("so far received ", len(allResponse)))
		if len(allResponse) >= len(endPoints) {
			Logger.Info("completed ...")

			break
		}

		Logger.Info("waiting ...")
		time.Sleep(time.Duration(1) * time.Second)
	}
	Logger.Info("responding with", zap.Any("all urls", allResponse))
	if contentBytes, errorMars := json.Marshal(allResponse); errorMars != nil {
		Logger.Error("error while marhalling", zap.Any("content", allResponse), zap.Error(errorMars))
	} else {
		w.Write([]byte(contentBytes))
	}

}
func init() {

	config := zap.NewDevelopmentConfig()
	config.Level.SetLevel(zap.InfoLevel)
	if Logger, crLoggErr = config.Build(); crLoggErr != nil {
		panic("error when setup the log")
	}
	flag.StringVar(&configPath, "config", "config/sam-ping.yaml", "a config string variable")
	flag.Parse()
	viper.SetConfigFile(configPath)
	if errRead := viper.ReadInConfig(); errRead != nil {
		Logger.Error("error while loading config", zap.String("config file", configPath), zap.Error(errRead))
	}
	Port = viper.GetInt("port")

	anEndpoints := viper.Get("endPoints")
	var ok bool
	if endPoints, ok = anEndpoints.([]interface{}); ok {
		Logger.Info("Reading configuration", zap.Int("port", Port), zap.Any("endpoints", endPoints))

	}

}
