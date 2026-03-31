package locale

import (
	"log"
	"os"
	"path"
	"runtime"

	"github.com/kweaver-ai/kweaver-go-lib/i18n"
)

var localeDir = "/locale"

func Register() {
	var abPath string

	// Auto-detect test mode: use UT mode if locale directory doesn't exist
	isTestMode := os.Getenv("I18N_MODE_UT") == "true"
	if !isTestMode {
		abPath, _ = os.Getwd()
		abPath += localeDir

		if _, err := os.Stat(abPath); os.IsNotExist(err) {
			// Locale directory doesn't exist, assume we're in test mode
			isTestMode = true
		}
	}

	// UT MODE
	if isTestMode {
		_, filename, _, ok := runtime.Caller(0)
		if ok {
			abPath = path.Dir(filename)
		} else {
			log.Println("locale: failed to get absolute path")
			return
		}
	} else {
		abPath, _ = os.Getwd()
		abPath += localeDir
	}

	i18n.RegisterI18n(abPath)
}
