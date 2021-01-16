package logger

import (
	"os"
	"fmt"
	"time"
)

func Debug(err error) {
	f, _ := os.OpenFile("octavia-driver-agent.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	f.WriteString(fmt.Sprintf("%s: %v\n",time.Now().Format(time.UnixDate),err.Error()))
}
