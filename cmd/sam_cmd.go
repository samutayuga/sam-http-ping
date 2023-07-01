package cmd

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/common-nighthawk/go-figure"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

var RootCommand = &cobra.Command{
	Use:   "RootCommand",
	Short: "The entry point for any command",
	Long: `This is a command based on cobra command
	`,
	TraverseChildren: true,
}

func LaunchHttpServer(componentName string) {
	router := mux.NewRouter()
	router.HandleFunc("/ping", RequestHandler)
	router.HandleFunc("/propagate", Propagate)
	address := fmt.Sprintf(":%d", Port)
	//get app name
	//appName := os.Getenv("APP_NAME")

	go func() {
		//start in another go routing
		Logger.Info("starting http server", zap.String("appName", componentName), zap.String("address", address))
		//go-figure.
		aFig := figure.NewColorFigure(componentName, "", "white", true)
		aFig.Print()

		if errListen := http.ListenAndServe(address, router); errListen != nil {
			Logger.Error("Error starting http server", zap.String("address", address), zap.Error(errListen))
		}

	}()
	//log.Fatal(http.ListenAndServe(address, router))
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch

	Logger.Info("Stopping server")
}

type HttpCmdParam struct {
	appName        string
	configLocation string
}
type KubeClientCmd struct {
	inCluster  string
	rConfig    *rest.Config
	kClientSet *kubernetes.Clientset
}

var LaunchHttpCommand = &cobra.Command{
	Use:   "launchHttp",
	Short: "Launch Http Server",
	Long:  `The command to launch the http server`,
	Args: func(cmd *cobra.Command, args []string) error {
		Logger.Info("args", zap.Any("args", args))
		return nil
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		viper.SetConfigFile(httpCmdParams.configLocation)
		if errRead := viper.ReadInConfig(); errRead != nil {
			Logger.Error("error while loading config", zap.String("config file", configPath), zap.Error(errRead))
		}
		Port = viper.GetInt("port")

		anEndpoints := viper.Get("endPoints")
		var ok bool
		if endPoints, ok = anEndpoints.([]interface{}); ok {
			Logger.Info("Reading configuration", zap.Int("port", Port), zap.Any("endpoints", endPoints))
			filteredEndPoints = make([]interface{}, 0)
			for _, aVal := range endPoints {
				if aMapEp, isMatch := aVal.(map[string]interface{}); isMatch {
					if aNamestr, isString := aMapEp["name"].(string); isString && aNamestr != AppName {
						filteredEndPoints = append(filteredEndPoints, aMapEp)
					}
				}
			}
			Logger.Info("Final endpoints", zap.Any("filteredEndPoints", filteredEndPoints))

		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		Logger.Info("args", zap.Any("args", args), zap.String("params", httpCmdParams.appName))
		LaunchHttpServer(httpCmdParams.appName)
		return nil
	},
}

var monitorDeployment = &cobra.Command{
	Use:   "monitorDeployment",
	Short: "Monitor Deployment",
	Long: `
	Monitor Deployment
	`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		//clientcmd.BuildConfigFromFlags
		Logger.Info("preRun for kubeclient", zap.Any("inCluster", kClientCmdParams.inCluster))
		var errConfig error
		if kClientCmdParams.inCluster == "false" {
			hDir := os.Getenv("HOME")
			fPath := filepath.Join(hDir, ".kube", "config")
			kClientCmdParams.rConfig, errConfig = clientcmd.BuildConfigFromFlags("", fPath)

		} else {
			kClientCmdParams.rConfig, errConfig = clientcmd.BuildConfigFromFlags("", "")

		}
		if errConfig != nil {
			Logger.Fatal("error while getting config for kubernetes", zap.Error(errConfig))
			return errConfig

		} else {

			Logger.Info("kube config is created", zap.String("config", kClientCmdParams.rConfig.Host))
			kClientCmdParams.kClientSet, errConfig = kubernetes.NewForConfig(kClientCmdParams.rConfig)
			if errConfig != nil {
				Logger.Fatal("error while creating clientset", zap.Error(errConfig))
				return errConfig
			}
			Logger.Info("successfully create clientset")

			return nil
		}

	},
	RunE: func(cmd *cobra.Command, args []string) error {
		stopper := make(chan struct{})
		defer close(stopper)
		defer runtime.HandleCrash()
		depInformerFactory := informers.NewSharedInformerFactory(kClientCmdParams.kClientSet, 0)
		informers := depInformerFactory.Apps().V1().Deployments().Informer()

		go depInformerFactory.Start(stopper)

		if !cache.WaitForCacheSync(stopper, informers.HasSynced) {
			runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
			return nil
		}

		informers.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				aDep := obj.(*v1.Deployment)
				Logger.Info("created deployment", zap.String("name", aDep.GetName()))

			},
			DeleteFunc: func(obj interface{}) {
				aDep := obj.(*v1.Deployment)
				Logger.Info("deleted deployment", zap.String("name", aDep.GetName()))
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				aDep := newObj.(*v1.Deployment)
				Logger.Info("updated deployment", zap.String("name", aDep.GetName()))
			},
		})

		<-stopper

		return nil
	},
}

func Execute() error {
	return RootCommand.Execute()
}
