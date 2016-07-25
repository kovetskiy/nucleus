package main

import (
	"fmt"
	"strings"
)

const (
	loggerFormatLength = 28
)

func fatalf(format string, values ...interface{}) {
	logger.Fatal(wrapLines(format, values...))
}

func errorf(format string, values ...interface{}) {
	logger.Error(wrapLines(format, values...))
}

func warningf(format string, values ...interface{}) {
	logger.Warning(wrapLines(format, values...))
}

func infof(format string, values ...interface{}) {
	logger.Info(wrapLines(format, values...))
}

func debugf(format string, values ...interface{}) {
	logger.Debug(wrapLines(format, values...))
}

func tracef(format string, values ...interface{}) {
	logger.Trace(wrapLines(format, values...))
}

func debugln(value interface{}) {
	logger.Debug(wrapLines("%s", value))
}

func infoln(value interface{}) {
	logger.Info(wrapLines("%s", value))
}

func fatalln(value interface{}) {
	logger.Fatal(wrapLines("%s", value))
}

func errorln(value interface{}) {
	logger.Error(wrapLines("%s", value))
}

func wrapLines(format string, values ...interface{}) string {
	contents := fmt.Sprintf(format, values...)
	contents = strings.TrimSuffix(contents, "\n")
	contents = strings.Replace(
		contents,
		"\n",
		"\n"+strings.Repeat(" ", loggerFormatLength),
		-1,
	)

	return contents
}
