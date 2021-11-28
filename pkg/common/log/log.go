package log

import (
	"bytes"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"strings"
)

var (
	// colorStatus returns a new function that returns status-colorized (cyan) strings for the
	// given arguments with fmt.Sprint().
	colorStatus = color.New(color.FgCyan).SprintFunc()

	// colorWarn returns a new function that returns status-colorized (yellow) strings for the
	// given arguments with fmt.Sprint().
	colorWarn = color.New(color.FgYellow).SprintFunc()

	// colorInfo returns a new function that returns info-colorized (green) strings for the
	// given arguments with fmt.Sprint().
	colorInfo = color.New(color.FgGreen).SprintFunc()

	// colorError returns a new function that returns error-colorized (red) strings for the
	// given arguments with fmt.Sprint().
	colorError = color.New(color.FgRed).SprintFunc()

	colorCommand = color.New(color.FgBlue).SprintFunc()

	logger *logrus.Entry

	labelsPath = "/etc/labels"
)

var ( // For Test Mocks
	initLogger = initializeLogger
)

var defaultLogger *logrus.Logger

// FormatLayoutType the layout kind
type FormatLayoutType string

// CustomTextFormat lets use a custom text format
type CustomTextFormat struct {
	ShowInfoLevel   bool
	ShowTimestamp   bool
	ShowSubCommand  string
	TimestampFormat string
}

func BeginSubCommandLogging(c string) {
	logrus.SetFormatter(NewCustomTextFormat(c))
}

func EndSubCommandLogging() {
	logrus.SetFormatter(NewCustomTextFormat(""))
}
func NewCustomTextFormat(cmd string) *CustomTextFormat {
	return &CustomTextFormat{
		ShowInfoLevel:   false,
		ShowTimestamp:   false,
		ShowSubCommand:  cmd,
		TimestampFormat: "2006-01-02 15:04:05",
	}
}

// Format formats the log statement
func (f *CustomTextFormat) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer

	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	level := strings.ToUpper(entry.Level.String())
	messageSplit := strings.Split(entry.Message, "\n")
	for _, message := range messageSplit {

		switch level {
		case "INFO":
			b.WriteString(colorInfo("INFO "))
			b.WriteString(": ")
		case "WARNING":
			b.WriteString(colorWarn("WARN "))
			b.WriteString(": ")
		case "DEBUG":
			b.WriteString(colorStatus("DEBUG"))
			b.WriteString(": ")
		case "ERROR":
			b.WriteString(colorError("ERROR"))
			b.WriteString(": ")
		case "FATAL":
			b.WriteString(colorError("FATAL"))
			b.WriteString(": ")
		default:
			b.WriteString(colorError(level))
			b.WriteString(": ")
		}
		if f.ShowSubCommand != "" {
			b.WriteString(colorCommand(strings.ToUpper(f.ShowSubCommand)))
			b.WriteString(" : ")
		}
		if f.ShowTimestamp {
			b.WriteString(entry.Time.Format(f.TimestampFormat))
			b.WriteString(" - ")
		}

		b.WriteString(message)

		if !strings.HasSuffix(message, "\n") {
			b.WriteByte('\n')
		}
	}
	return b.Bytes(), nil
}

func initializeLogger() error {
	if logger == nil {
		var fields logrus.Fields
		logger = logrus.WithFields(fields)

		format := os.Getenv("LOG_FORMAT")
		if format == "json" {
			setFormatter("json")
		} else {
			setFormatter("text")
		}
	}
	return nil
}

// setFormatter sets the logrus format to use either text or JSON formatting
func setFormatter(layout FormatLayoutType) {
	switch layout {
	case "json":
		logrus.SetFormatter(&logrus.JSONFormatter{})
	default:
		logrus.SetFormatter(NewCustomTextFormat(""))
	}
}

// Logger obtains the logger for use in the codebase
// This is the only way you should obtain a logger
func Logger() *logrus.Entry {
	err := initLogger()
	if err != nil {
		logrus.Warnf("error initializing logrus %v", err)
	}
	return logger
}

// SetLevel sets the logging level
func SetLevel(s string) error {
	level, err := logrus.ParseLevel(s)
	if err != nil {
		return errors.Errorf("Invalid log level '%s'", s)
	}
	logrus.SetLevel(level)
	return nil
}

// CaptureOutput calls the specified function capturing and returning all logged messages.
func CaptureOutput(f func()) string {
	var buf bytes.Buffer
	logrus.SetOutput(&buf)
	f()
	logrus.SetOutput(os.Stderr)
	return buf.String()
}

// SetOutput sets the outputs for the default logger.
func SetOutput(out io.Writer) {
	logrus.SetOutput(out)
}

// GetLevels returns the list of valid log levels
func GetLevels() []string {
	var levels []string
	for _, level := range logrus.AllLevels {
		levels = append(levels, level.String())
	}
	return levels
}
