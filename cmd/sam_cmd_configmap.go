package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var monitorConfigMap = &cobra.Command{
	Use:   "monitorConfigMap",
	Short: "Monitor ConfigMap",
	Long: `
	Monitor ConfigMap
	`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		//clientcmd.BuildConfigFromFlags
		Logger.Info("preRun for kubeclient", zap.Any("inCluster", kClientCmdParams.inCluster))
		var errConfig error
		if kClientCmdParams.kClientSet, errConfig = buildKubeClient(); errConfig != nil {
			return errConfig
		}
		return nil

	},

	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}
