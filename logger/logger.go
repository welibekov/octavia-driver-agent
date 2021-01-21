package logger

import (
	"os"
	"fmt"
	"time"
)
var LogFile string

func Debug(err error) {
	f, _ := os.OpenFile(LogFile,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	f.WriteString(fmt.Sprintf("%s: %v\n",time.Now().Format(time.UnixDate),err.Error()))
}
