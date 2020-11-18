package support

import (
	"fmt"
	"os"
	"time"
)

func ErrorLog(err error) {
	errorlog, rip := os.OpenFile("error.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	// If we encounter an error here, something is seriously wrong.
	if rip != nil {
		return
	}
	defer errorlog.Close()

	t := time.Now()
	date := fmt.Sprintf("%02d-%02d-%04d_%02d-%02d-%02d", t.Month(), t.Day(), t.Year(), t.Hour(), t.Minute(), t.Second())

	errorlog.WriteString(fmt.Sprintf("%s: %s\n", date, err))

	return
}

func Log(text string) {
	log, rip := os.OpenFile("react.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	// If we encounter an error here, something is seriously wrong.
	if rip != nil {
		return
	}
	defer log.Close()

	t := time.Now()
	date := fmt.Sprintf("%02d-%02d-%04d_%02d-%02d-%02d", t.Month(), t.Day(), t.Year(), t.Hour(), t.Minute(), t.Second())

	log.WriteString(fmt.Sprintf("%s: %s\n", date, text))

	return
}
