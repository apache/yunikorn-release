package setup

import (
	"github.com/apache/yunikorn-release/soak/framework"
	"go.uber.org/zap"
	"testing"
)

func TestSetAutoScalerPerConfig(t *testing.T) {
	conf, err := framework.InitConfig("test_conf.yaml")
	if err != nil {
		logger.Fatal("failed to parse config", zap.Error(err))
	}
	logger.Info("config successfully loaded", zap.Any("conf", conf))

	for _, test := range conf.Tests {
		if len(test.Template.Node) > 0 {
			for _, nodeTemplate := range test.Template.Node {
				err := setAutoscalerPerConfig(nodeTemplate)
				if err != nil {
					t.Fatal(err)
				}
			}
		}
	}
}
