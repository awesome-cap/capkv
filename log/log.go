/**
2 * @Author: Nico
3 * @Date: 2021/6/10 0:27
4 */
package log

import (
	"fmt"
	"os"
)

func write(level string, str string){
	fmt.Printf("[%s] %s", level, str)
}

func Info(str interface{}){
	write("INFO", fmt.Sprintf("%v", str))
}

func Infof(format string, args ...interface{}){
	write("INFO", fmt.Sprintf(format, args...))
}

func Debugf(format string, args ...interface{}){
	write("DEBUG", fmt.Sprintf(format, args...))
}

func Errorf(format string, args ...interface{}){
	write("ERROR", fmt.Sprintf(format, args...))
}

func Warnf(format string, args ...interface{}){
	write("WARN", fmt.Sprintf(format, args...))
}

func Fatal(str interface{}){
	write("FATAL", fmt.Sprintf("%v", str))
	os.Exit(0)
}

func Fatalf(format string, args ...interface{}){
	write("FATAL", fmt.Sprintf(format, args...))
	os.Exit(0)
}