package locale

import (
	"log"
	"os"
	"path"
	"runtime"

	"github.com/kweaver-ai/kweaver-go-lib/i18n"
)

var (
	localeDir = "/locale"
)

func Register() {
	var abPath string

	// UT MODE
	if os.Getenv("I18N_MODE_UT") == "true" {
		_, filename, _, ok := runtime.Caller(0)
		if ok {
			abPath = path.Dir(filename)
		} else {
			log.Fatal("failed to get absolute path")
		}
	} else {
		abPath, _ = os.Getwd()
		abPath += localeDir
	}
	i18n.RegisterI18n(abPath)
}
