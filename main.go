package main

import (
	"sam-http-ping/cmd"

	"go.uber.org/zap"
)

func main() {
	if err := cmd.Execute(); err != nil {
		cmd.Logger.Error("error while launching command", zap.Error(err))
	}

}
