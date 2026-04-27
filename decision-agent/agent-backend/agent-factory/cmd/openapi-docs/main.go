package main

import "os"

// main 根据子命令分发到 generate、compare、validate 三条工作流。
func main() {
	if len(os.Args) < 2 {
		exitWithError(usageError())
	}

	var err error

	switch os.Args[1] {
	case "generate":
		err = runGenerate(os.Args[2:])
	case "compare":
		err = runCompare(os.Args[2:])
	case "validate":
		err = runValidate(os.Args[2:])
	default:
		err = usageError()
	}

	if err != nil {
		exitWithError(err)
	}
}
