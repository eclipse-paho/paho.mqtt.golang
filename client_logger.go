package mqtt

// clientLogger provides a structured way to log messages for different client actions.
// It contains loggers for ERROR, CRITICAL, WARN, and DEBUG levels, each associated with a specific client ID.
type clientLogger struct {
	ERROR    clientErrorLogger
	CRITICAL clientCriticalLogger
	WARN     clientWarnLogger
	DEBUG    clientDebugLogger
}

// clientErrorLogger is a wrapper around a Logger that prefixes log messages with the client ID.
// It implements the Logger interface, allowing it to be used interchangeably with other loggers.
type clientErrorLogger struct {
	clientID string
	logger   Logger
}

func (cel *clientErrorLogger) Println(v ...interface{}) {
	v = append([]interface{}{"[" + cel.clientID + "]"}, v...)
	cel.logger.Println(v...)
}

func (cel *clientErrorLogger) Printf(format string, v ...interface{}) {
	format = "[" + cel.clientID + "] " + format
	cel.logger.Printf(format, v...)
}

// clientCriticalLogger is a wrapper around a Logger that prefixes critical log messages with the client ID.
// It implements the Logger interface, allowing it to be used interchangeably with other loggers.
type clientCriticalLogger struct {
	clientID string
	logger   Logger
}

func (ccl *clientCriticalLogger) Println(v ...interface{}) {
	v = append([]interface{}{"[" + ccl.clientID + "]"}, v...)
	ccl.logger.Println(v...)
}

func (ccl *clientCriticalLogger) Printf(format string, v ...interface{}) {
	format = "[" + ccl.clientID + "] " + format
	ccl.logger.Printf(format, v...)
}

// clientWarnLogger is a wrapper around a Logger that prefixes warning log messages with the client ID.
// It implements the Logger interface, allowing it to be used interchangeably with other loggers.
type clientWarnLogger struct {
	clientID string
	logger   Logger
}

func (cwl *clientWarnLogger) Println(v ...interface{}) {
	v = append([]interface{}{"[" + cwl.clientID + "]"}, v...)
	cwl.logger.Println(v...)
}

func (cwl *clientWarnLogger) Printf(format string, v ...interface{}) {
	format = "[" + cwl.clientID + "] " + format
	cwl.logger.Printf(format, v...)
}

// clientDebugLogger is a wrapper around a Logger that prefixes debug log messages with the client ID.
// It implements the Logger interface, allowing it to be used interchangeably with other loggers.
type clientDebugLogger struct {
	clientID string
	logger   Logger
}

func (cdl *clientDebugLogger) Println(v ...interface{}) {
	v = append([]interface{}{"[" + cdl.clientID + "]"}, v...)
	cdl.logger.Println(v...)
}

func (cdl *clientDebugLogger) Printf(format string, v ...interface{}) {
	format = "[" + cdl.clientID + "] " + format
	cdl.logger.Printf(format, v...)
}

// NewClientLogger creates a new clientLogger with the provided clientID and loggers for different levels.
// If any of the loggers are nil, they will default to NOOPLogger.
func NewClientLogger(clientID string, errorLogger, criticalLogger, warnLogger, debugLogger Logger) *clientLogger {
	if errorLogger == nil {
		errorLogger = NOOPLogger{}
	}
	if criticalLogger == nil {
		criticalLogger = NOOPLogger{}
	}
	if warnLogger == nil {
		warnLogger = NOOPLogger{}
	}
	if debugLogger == nil {
		debugLogger = NOOPLogger{}
	}
	return &clientLogger{
		ERROR:    clientErrorLogger{clientID: clientID, logger: errorLogger},
		CRITICAL: clientCriticalLogger{clientID: clientID, logger: criticalLogger},
		WARN:     clientWarnLogger{clientID: clientID, logger: warnLogger},
		DEBUG:    clientDebugLogger{clientID: clientID, logger: debugLogger},
	}
}
