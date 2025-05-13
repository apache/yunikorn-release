package framework

import (
	"github.com/apache/yunikorn-release/soak/logger"
	"go.uber.org/zap"
)

var log = logger.Logger

type Scenarios struct {
	registeredTestScenarios map[string]TestScenario
}

var testScenarios Scenarios

func init() {
	testScenarios.registeredTestScenarios = make(map[string]TestScenario)
}

func Register(ts TestScenario) {
	testScenarios.registeredTestScenarios[ts.GetName()] = ts
	log.Info("register scenario", zap.String("scenarioName", ts.GetName()))
}

func GetRegisteredTestScenarios() map[string]TestScenario {
	return testScenarios.registeredTestScenarios
}

type TestScenario interface {
	GetName() string
	Init() error
	Run() ([]string, error)
}
