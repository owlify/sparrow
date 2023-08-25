package environment

import "os"

type Environment string

type EnvProvider interface {
	CurrEnv() Environment
}

type osEnv struct {
	key string
}

func NewOsEnv() *osEnv {
	return &osEnv{key: "ENV"}
}

func (e Environment) String() string {
	return string(e)
}

func (e *osEnv) CurrEnv() Environment {
	env := Environment(os.Getenv(e.key))
	switch env {
	case DevEnv, TestingEnv, StagingEnv, QAEnv, SandboxEnv, ProductionEnv, UnicornEnv:
		return env
	default:
		panic("current os env is not a valid value")
	}
}

func (e Environment) IsStructuredLogging() bool {
	switch e {
	case ProductionEnv, SandboxEnv, StagingEnv, QAEnv, UnicornEnv:
		return true
	default:
		return false
	}
}
