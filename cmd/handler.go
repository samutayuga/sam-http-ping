package cmd

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

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
	DnsCheckingMsg  string
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

		r := SamResponse{ResponseCode: httpResp.StatusCode, ResponseMessage: httpResp.Status, Origin: hostN, Destination: url}
		//Logger.Info("Response with", zap.Int("response", r.ResponseCode))
		return &r
	} else {
		Logger.Warn("failed the make http call", zap.String("url", url), zap.Error(errGet))
		return &SamResponse{ResponseCode: -1, ResponseMessage: errGet.Error(), Origin: hostN, Destination: url}
	}

}

var (
	Logger              *zap.Logger
	crLoggErr           error
	configPath, AppName string
	endPoints           []interface{}
	Port                int
	filteredEndPoints   []interface{}
	kClientCmdParams    KubeClientCmd
	httpCmdParams       HttpCmdParam
	//givenAppName        string
	//httpParam           *HttpCmdParam
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

func resolveIP(aVal string) string {
	sUrl, parseUrlErr := url.ParseRequestURI(aVal)
	if parseUrlErr != nil {
		Logger.Error("cannot parse url string", zap.String("url", aVal), zap.Error(parseUrlErr))
	}
	aHost := sUrl.Host
	aHostArr := strings.Split(aHost, ":")

	ipAddress, errGetIp := net.LookupIP(aHostArr[0])

	if errGetIp != nil {
		Logger.Error("cannot resolve ip address", zap.String("host", aHost), zap.Error(errGetIp))
		return errGetIp.Error()
	}
	Logger.Info("ip address is resolved properly", zap.String("host", aHost), zap.Any("ip address", ipAddress))
	return fmt.Sprintf("%s is resolved succesfully, ip address %v ", aHost, ipAddress)

}

func doPropagate(singleMap interface{}, payload *SamPayload, resp chan *SamResponse) {
	if aMap, correct := singleMap.(map[string]interface{}); correct {
		if aVal, exists := aMap["url"]; exists {
			//resolve ip
			ipResolution := resolveIP(aVal.(string))

			aResp := payload.doGet(aVal.(string))
			aResp.DnsCheckingMsg = ipResolution

			Logger.Info("response from get url ", zap.String("url", aVal.(string)),
				zap.Int("responseCode", aResp.ResponseCode),
				zap.String("responseMessage", aResp.ResponseMessage))

			resp <- aResp

		}
	}
	//return nil
}
func managePropagation(requestor *SamPayload, allResponseChan chan *SamResponse) {
	for _, _val := range filteredEndPoints {
		go func(myVal interface{}) {
			doPropagate(myVal, requestor, allResponseChan)
		}(_val)

	}
}
func Propagate(w http.ResponseWriter, r *http.Request) {
	Logger.Info("Serving request", zap.String("origin", r.Host))
	p := SamPayload{httClient: &http.Client{}}

	allResponse := make([]*SamResponse, 0)
	allResponseChan := make(chan *SamResponse, len(filteredEndPoints))
	managePropagation(&p, allResponseChan)
	//monitor the response
	for {
		allResponse = append(allResponse, <-allResponseChan)
		Logger.Info("receive result", zap.Int("so far received ", len(allResponse)))
		if len(allResponse) >= len(filteredEndPoints) {
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

	//AppName = os.Getenv("APP_NAME")
	config := zap.NewDevelopmentConfig()
	config.Level.SetLevel(zap.InfoLevel)
	if Logger, crLoggErr = config.Build(); crLoggErr != nil {
		panic("error when setup the log")
	}
	// flag.StringVar(&configPath, "config", "config/sam-ping.yaml", "a config string variable")
	// flag.Parse()
	// viper.SetConfigFile(configPath)
	// if errRead := viper.ReadInConfig(); errRead != nil {
	// 	Logger.Error("error while loading config", zap.String("config file", configPath), zap.Error(errRead))
	// }
	// Port = viper.GetInt("port")

	// anEndpoints := viper.Get("endPoints")
	// var ok bool
	// if endPoints, ok = anEndpoints.([]interface{}); ok {
	// 	Logger.Info("Reading configuration", zap.Int("port", Port), zap.Any("endpoints", endPoints))
	// 	filteredEndPoints = make([]interface{}, 0)
	// 	for _, aVal := range endPoints {
	// 		if aMapEp, isMatch := aVal.(map[string]interface{}); isMatch {
	// 			if aNamestr, isString := aMapEp["name"].(string); isString && aNamestr != AppName {
	// 				filteredEndPoints = append(filteredEndPoints, aMapEp)
	// 			}
	// 		}
	// 	}
	// 	Logger.Info("Final endpoints", zap.Any("filteredEndPoints", filteredEndPoints))

	// }

	//httpParam = HttpCmdParam{}

	kClientCmdParams = KubeClientCmd{
		inCluster: "",
	}
	httpCmdParams = HttpCmdParam{}

	RootCommand.AddCommand(LaunchHttpCommand)
	LaunchHttpCommand.Flags().StringVarP(&httpCmdParams.appName, "appName", "a", "backend", "launchHttp -a backend")
	LaunchHttpCommand.Flags().StringVarP(&httpCmdParams.configLocation, "config", "c", "config/sam-ping.yaml", "launchHttp -c config/sam-ping.yaml")
	LaunchHttpCommand.MarkFlagRequired("appName")
	LaunchHttpCommand.MarkFlagRequired("config")
	RootCommand.AddCommand(monitorDeployment)
	monitorDeployment.Flags().StringVarP(&kClientCmdParams.inCluster, "inCluster", "i", "false", "monitorDeployment -i false")
	monitorDeployment.MarkFlagRequired("inCluster")

}
