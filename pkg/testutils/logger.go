package testutils

import (
	"os"

	"github.com/gnzdotmx/ishinobu/pkg/utils"
)

func NewTestLogger() *utils.Logger {
	return utils.NewLoggerWithFile(os.Stdout)
}
