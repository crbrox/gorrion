// +build !debug

package gorrion

type loggerT struct{}

func (loggerT) Printf(format string, v ...interface{}) {}
func (loggerT) Println(v ...interface{})               {}
func (loggerT) Print(v ...interface{})                 {}

func init() {
	logger = loggerT{}
}
