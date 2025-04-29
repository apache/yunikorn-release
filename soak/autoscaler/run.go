package autoscaler

import (
	"fmt"
	"github.com/apache/yunikorn-release/soak/framework"
	"github.com/apache/yunikorn-release/soak/logger"
	"go.uber.org/zap"
	"os/exec"
	"path/filepath"
	"strings"
)

var log = logger.Logger

type AutoscalingScenario struct {
	templateConf framework.Template
	testCases    []framework.TestCase
}

func New(config *framework.Config) *AutoscalingScenario {
	for _, c := range config.Tests {
		if c.Name == "autoscaling" {
			return &AutoscalingScenario{
				templateConf: c.Template,
				testCases:    c.TestCases,
			}
		}
	}
	return nil
}

func (a *AutoscalingScenario) GetName() string {
	return "autoscaling"
}

func (a *AutoscalingScenario) Init() error {
	if err := a.upgradeSchedulerPerConfig(); err != nil {
		return err
	}

	return a.setAutoscalerPerConfig()
}

func (a *AutoscalingScenario) Tests() []framework.TestCase {
	// enable or disable test cases here
	return a.testCases
}

func (a *AutoscalingScenario) Run() ([]string, error) {
	log := logger.Logger
	results := make([]string, len(a.testCases))
	for idx, tests := range a.testCases {
		clusterLoaderConfigPath := tests.ClusterLoaderConfigPath
		reportDir := filepath.Dir(clusterLoaderConfigPath)
		args := []string{fmt.Sprintf("----testconfig=%s", clusterLoaderConfigPath),
			"--provider=kind", fmt.Sprintf("--kubeconfig=%s", a.templateConf.Kubeconfig.Path),
			"--v=4", fmt.Sprintf("--report-dir=%s", reportDir)}
		cmd := exec.Command("clusterloader2", args...)
		log.Info("Clusterloader command to be executed",
			zap.String("command", fmt.Sprintf("clusterloader2 %s", strings.Join(args, " "))))
		results[idx] = reportDir
		_, err := cmd.CombinedOutput()
		if err != nil {
			log.Error("Clusterloader command failed. Check results directory for more info",
				zap.String("command", fmt.Sprintf("clusterloader2 %s", strings.Join(args, " "))))

			return results, err

		}
	}
	return results, nil
}
