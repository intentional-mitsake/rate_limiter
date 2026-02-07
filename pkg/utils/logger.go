package utils

import (
	"log"
	"os"
)

// three types of logs we will be printing
// may add more later
type Logger struct {
	//*log.Logger is a pointer to a log.Logger struct from the standard library
	//it is used to log info, error, and warning messages respectively
	//the looger object can be used to write log messages to different outputs
	iLog *log.Logger
	eLog *log.Logger
	wLog *log.Logger
}

func CreateLogger() *Logger {
	return &Logger{
		//log.New is a function from the standard library that creates a new logger
		//it takes an output writer, a prefix string, and a flag value
		//here os.Stdout and os.Stderr are used as output writers
		//stderr is diff from stdout in that it is used specifically for error messages
		//prefixes are "INFO: ", "ERROR: ", and "WARNING: "
		//log.LstdFlags is a flag that adds the date and time to each log message
		iLog: log.New(os.Stdout, "INFO: ", log.LstdFlags),
		eLog: log.New(os.Stderr, "ERROR: ", log.LstdFlags),
		wLog: log.New(os.Stdout, "WARNING: ", log.LstdFlags),
	}
}

// Info method works on Logger struct to log info messages and can only be called on Logger instances
// it takes a message string as input and uses the iLog logger to print the message
func (l *Logger) Info(msg string) {
	//use the iLog logger to print the info message
	l.iLog.Println(msg)
}

func (l *Logger) Error(msg string) {
	//use the eLog logger to print the error message
	l.eLog.Println(msg)
}
func (l *Logger) Warning(msg string) {
	//use the wLog logger to print the warning message
	l.wLog.Println(msg)
}
