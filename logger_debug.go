// +build debug

package gorrion

import (
	"log"
	"os"
)

type loggerT struct {
	*log.Logger
}

func init() {
	logger = loggerT{log.New(os.Stdout, "", log.LstdFlags)}
}
