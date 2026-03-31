package cenvhelper

// IsLocalDev 判断是否为本地开发环境
// runScenarios: 运行场景
//   - 如果 runScenarios 为空， 则判断是否为本地开发环境
//   - 如果 runScenarios 不为空， 则判断是否为本地开发环境且运行场景在 runScenarios 中
func IsLocalDev(runScenarios ...RunScenario) (_isLocalDev bool) {
	if !isEnvInited {
		panic("env not inited")
	}

	// 1. 第一层判断
	if isLocalDev.Value() != "true" {
		return
	}

	// 2. 第二层判断

	if len(runScenarios) == 0 {
		_isLocalDev = true
		return
	}

	// 2.1 获取运行场景对应的环境变量
	runScenarioVal := runScenarioEnv.Value()

	if runScenarioVal == "" {
		return
	}

	// 2.2 获取运行场景对应的环境变量值
	for _, runScenario := range runScenarios {
		if runScenarioVal == runScenario.String() {
			_isLocalDev = true
			return
		}
	}

	return
}

func IsAaronLocalDev() bool {
	_isLocalDev := IsLocalDev(RunScenario_Aaron_Local_Dev)
	return _isLocalDev
}
