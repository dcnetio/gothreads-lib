// Package logger provides a simple logging package.
package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/fatih/color"
	logging "github.com/ipfs/go-log/v2"
)

const (
	LEVEL_NOSET = iota
	LEVEL_DEBUG
	LEVEL_INFO
	LEVEL_WARNING
	LEVEL_ERROR
	LEVEL_FATAL
	LEVEL_SUCCESS
)

// This is the set of emoji definition.
const (
	FatalEmoji   = "☠️"
	DebugEmoji   = "🐞"
	InfoEmoji    = "🦋"
	SuccessEmoji = "🎉"
	ErrorEmoji   = "💔"
	WarningEmoji = "⚠️"
)

// It contains configuration of logger.
var (

	// Level defines the level of the logger. It's a global option and affects
	Level = LEVEL_NOSET

	FilelineFlag = false
	//TimeStamps defines if the output has timestamps or not. It's a global option and affects all methods.
	TimeStamps = true
	// Color defines if the output has color or not. It's a global option and affects all methods.
	Color = true
	// BackgroundColor defines if the output has background color or not. It's a global option and affects all methods.
	BackgroundColor = false
)

// Debug formats according to a format specifier and writes to standard output.
func Debug(format string, a ...interface{}) {
	if Level <= LEVEL_DEBUG {
		a, w := extractLoggerArguments(format, a...)
		s := fmt.Sprintf(emojiLabel(format, DebugEmoji), a...)
		if Color && BackgroundColor {
			Color = false
		}
		if Color {
			c := color.New(color.FgHiBlue)
			c.EnableColor()
			s = c.Sprintf(s)
		}
		if BackgroundColor {
			c := color.New(color.FgHiWhite)
			c.EnableColor()
			s = c.Add(color.BgHiBlue).Sprintf(s)
		}
		prefix := generateLogPrefix(FilelineFlag, 2)
		content := fmt.Sprintf("%s %s", prefix, s)
		fmt.Fprint(w, content)
	}
}

// Info formats according to a format specifier and writes to standard output.
func Info(format string, a ...interface{}) {
	if Level <= LEVEL_INFO {
		a, w := extractLoggerArguments(format, a...)
		s := fmt.Sprintf(emojiLabel(format, InfoEmoji), a...)
		if Color && BackgroundColor {
			Color = false
		}
		if Color {
			c := color.New(color.FgHiCyan)
			c.EnableColor()
			s = c.Sprintf(s)
		}
		if BackgroundColor {
			c := color.New(color.FgHiWhite)
			c.EnableColor()
			s = c.Add(color.BgHiCyan).Sprintf(s)
		}
		prefix := generateLogPrefix(FilelineFlag, 2)
		content := fmt.Sprintf("%s %s", prefix, s)
		fmt.Fprint(w, content)
	}
}

// Success formats according to a format specifier and writes to standard output.
func Success(format string, a ...interface{}) {
	if Level <= LEVEL_SUCCESS {
		a, w := extractLoggerArguments(format, a...)
		s := fmt.Sprintf(emojiLabel(format, SuccessEmoji), a...)
		if Color && BackgroundColor {
			Color = false
		}
		if Color {
			c := color.New(color.FgHiGreen)
			c.EnableColor()
			s = c.Sprintf(s)

		}
		if BackgroundColor {
			c := color.New(color.FgHiWhite)
			c.EnableColor()
			s = c.Add(color.BgHiGreen).Sprintf(s)
		}
		prefix := generateLogPrefix(FilelineFlag, 2)
		content := fmt.Sprintf("%s %s", prefix, s)
		fmt.Fprint(w, content)
	}
}

// Warning formats according to a format specifier and writes to standard output.
func Warning(format string, a ...interface{}) {
	if Level <= LEVEL_WARNING {
		a, w := extractLoggerArguments(format, a...)
		s := fmt.Sprintf(emojiLabel(format, WarningEmoji), a...)
		if Color && BackgroundColor {
			Color = false
		}
		if Color {
			c := color.New(color.FgHiYellow)
			c.EnableColor()
			s = c.Sprintf(s)
		}
		if BackgroundColor {
			c := color.New(color.FgHiWhite)
			c.EnableColor()
			s = c.Add(color.BgHiYellow).Sprintf(s)
		}
		prefix := generateLogPrefix(FilelineFlag, 2)
		content := fmt.Sprintf("%s %s", prefix, s)
		fmt.Fprint(w, content)
	}
}

// Error formats according to a format specifier and writes to standard output.
func Error(format string, a ...interface{}) {
	if Level <= LEVEL_ERROR {
		a, w := extractLoggerArguments(format, a...)
		s := fmt.Sprintf(emojiLabel(format, ErrorEmoji), a...)
		if Color && BackgroundColor {
			Color = false
		}
		if Color {
			c := color.New(color.FgHiRed)
			c.EnableColor()
			s = c.Sprintf(s)
		}
		if BackgroundColor {
			c := color.New(color.FgHiWhite)
			c.EnableColor()
			s = c.Add(color.BgHiRed).Sprintf(s)
		}
		prefix := generateLogPrefix(FilelineFlag, 2)
		content := fmt.Sprintf("%s %s", prefix, s)
		fmt.Fprint(w, content)
	}
}

// Fatal formats according to a format specifier and writes to standard output.
func Fatal(format string, a ...interface{}) {
	if Level <= LEVEL_FATAL {
		a, w := extractLoggerArguments(format, a...)
		s := fmt.Sprintf(emojiLabel(format, FatalEmoji), a...)
		if Color && BackgroundColor {
			Color = false
		}
		if Color {
			c := color.New(color.FgRed)
			c.EnableColor()
			s = c.Sprintf(s)
		}
		if BackgroundColor {
			c := color.New(color.FgWhite)
			c.EnableColor()
			s = c.Add(color.BgHiRed).Sprintf(s)
		}
		prefix := generateLogPrefix(FilelineFlag, 2)
		content := fmt.Sprintf("%s %s", prefix, s)
		fmt.Fprint(w, content)

	}
}

type CompatLogger struct {
	PrefixLevel  int
	Level        int
	FilelineFlag bool
}

func (l *CompatLogger) Infof(format string, a ...interface{}) {
	if l.Level <= LEVEL_INFO {
		a, w := extractLoggerArguments(format, a...)
		s := fmt.Sprintf(emojiLabel(format, InfoEmoji), a...)
		if Color && BackgroundColor {
			Color = false
		}
		if Color {
			c := color.New(color.FgHiCyan)
			c.EnableColor()
			s = c.Sprintf(s)
		}
		if BackgroundColor {
			c := color.New(color.FgHiWhite)
			c.EnableColor()
			s = c.Add(color.BgHiCyan).Sprintf(s)
		}
		prefixLevel := 3
		if l.PrefixLevel > 0 {
			prefixLevel = l.PrefixLevel
		}
		prefix := generateLogPrefix(l.FilelineFlag, prefixLevel)
		content := fmt.Sprintf("%s %s", prefix, s)
		fmt.Fprint(w, content)
	}
}

// Debug formats according to a format specifier and writes to standard output.
func (l *CompatLogger) Debugf(format string, a ...interface{}) {
	if l.Level <= LEVEL_DEBUG {
		a, w := extractLoggerArguments(format, a...)
		s := fmt.Sprintf(emojiLabel(format, DebugEmoji), a...)
		if Color && BackgroundColor {
			Color = false
		}
		if Color {
			c := color.New(color.FgHiBlue)
			c.EnableColor()
			s = c.Sprintf(s)
		}
		if BackgroundColor {
			c := color.New(color.FgHiWhite)
			c.EnableColor()
			s = c.Add(color.BgHiBlue).Sprintf(s)
		}
		prefixLevel := 3
		if l.PrefixLevel > 0 {
			prefixLevel = l.PrefixLevel
		}
		prefix := generateLogPrefix(l.FilelineFlag, prefixLevel)
		content := fmt.Sprintf("%s %s", prefix, s)
		fmt.Fprint(w, content)
	}
}

// Success formats according to a format specifier and writes to standard output.
func (l *CompatLogger) Successf(format string, a ...interface{}) {
	if l.Level <= LEVEL_SUCCESS {
		a, w := extractLoggerArguments(format, a...)
		s := fmt.Sprintf(emojiLabel(format, SuccessEmoji), a...)
		if Color && BackgroundColor {
			Color = false
		}
		if Color {
			c := color.New(color.FgHiGreen)
			c.EnableColor()
			s = c.Sprintf(s)

		}
		if BackgroundColor {
			c := color.New(color.FgHiWhite)
			c.EnableColor()
			s = c.Add(color.BgHiGreen).Sprintf(s)
		}
		prefixLevel := 3
		if l.PrefixLevel > 0 {
			prefixLevel = l.PrefixLevel
		}
		prefix := generateLogPrefix(l.FilelineFlag, prefixLevel)
		content := fmt.Sprintf("%s %s", prefix, s)
		fmt.Fprint(w, content)
	}
}

// Warning formats according to a format specifier and writes to standard output.
func (l *CompatLogger) Warningf(format string, a ...interface{}) {
	if l.Level <= LEVEL_WARNING {
		a, w := extractLoggerArguments(format, a...)
		s := fmt.Sprintf(emojiLabel(format, WarningEmoji), a...)
		if Color && BackgroundColor {
			Color = false
		}
		if Color {
			c := color.New(color.FgHiYellow)
			c.EnableColor()
			s = c.Sprintf(s)
		}
		if BackgroundColor {
			c := color.New(color.FgHiWhite)
			c.EnableColor()
			s = c.Add(color.BgHiYellow).Sprintf(s)
		}
		prefixLevel := 3
		if l.PrefixLevel > 0 {
			prefixLevel = l.PrefixLevel
		}
		prefix := generateLogPrefix(l.FilelineFlag, prefixLevel)
		content := fmt.Sprintf("%s %s", prefix, s)
		fmt.Fprint(w, content)
	}
}

// Error formats according to a format specifier and writes to standard output.
func (l *CompatLogger) Errorf(format string, a ...interface{}) {
	if l.Level <= LEVEL_ERROR {
		a, w := extractLoggerArguments(format, a...)
		s := fmt.Sprintf(emojiLabel(format, ErrorEmoji), a...)
		if Color && BackgroundColor {
			Color = false
		}
		if Color {
			c := color.New(color.FgHiRed)
			c.EnableColor()
			s = c.Sprintf(s)
		}
		if BackgroundColor {
			c := color.New(color.FgHiWhite)
			c.EnableColor()
			s = c.Add(color.BgHiRed).Sprintf(s)
		}
		prefixLevel := 3
		if l.PrefixLevel > 0 {
			prefixLevel = l.PrefixLevel
		}
		prefix := generateLogPrefix(l.FilelineFlag, prefixLevel)
		content := fmt.Sprintf("%s %s", prefix, s)
		fmt.Fprint(w, content)
	}
}

// Fatal formats according to a format specifier and writes to standard output.
func (l *CompatLogger) Fatalf(format string, a ...interface{}) {
	if l.Level <= LEVEL_FATAL {
		a, w := extractLoggerArguments(format, a...)
		s := fmt.Sprintf(emojiLabel(format, FatalEmoji), a...)
		if Color && BackgroundColor {
			Color = false
		}
		if Color {
			c := color.New(color.FgRed)
			c.EnableColor()
			s = c.Sprintf(s)
		}
		if BackgroundColor {
			c := color.New(color.FgHiWhite)
			c.EnableColor()
			s = c.Add(color.BgRed).Sprintf(s)
		}
		prefixLevel := 3
		if l.PrefixLevel > 0 {
			prefixLevel = l.PrefixLevel
		}
		prefix := generateLogPrefix(l.FilelineFlag, prefixLevel)
		content := fmt.Sprintf("%s %s", prefix, s)
		fmt.Fprint(w, content)

	}
}

func extractLoggerArguments(format string, a ...interface{}) ([]interface{}, io.Writer) {
	n := strings.Count(format, "%")
	if n > len(a) {
		n = len(a)
	}
	var w io.Writer = os.Stdout
	if length := n; length > 0 {
		// extract an io.Writer at the end of a
		if value, ok := a[length-1].(io.Writer); ok {
			w = value
			a = a[0 : length-1]
		}
	}
	return a, w
}

func emojiLabel(format, label string) string {
	if !strings.Contains(format, "\n") {
		format = fmt.Sprintf("%s%s", format, "\n")
	}
	return fmt.Sprintf("%s-%s", label, format)
}

// generateLogPrefix generates a log prefix based on the current time and the file and line number
func generateLogPrefix(filelineFlag bool, prefixLevel int) string {
	filelinePrefix := ""
	if filelineFlag {
		// add file and line number
		_, file, line, ok := runtime.Caller(prefixLevel)
		if ok {
			//get short file name
			//	file = filepath.Base(file)
			filelinePrefix = fmt.Sprintf("%s:%d", file, line)
		}
	}
	if TimeStamps {
		t := time.Now()
		rfcTime := t.Format(time.RFC3339)
		return fmt.Sprintf("%s %s", rfcTime, filelinePrefix)
	} else {
		return filelinePrefix
	}
}

// InitLogger initializes the logger
func InitAndCloseOther() {
	c := logging.Config{
		Format: logging.ColorizedOutput,
		Stdout: false,
		Stderr: true,
		Level:  logging.LevelError,
	}
	logging.SetupLogging(c)
	logging.SetLogLevel("*", "ERROR")

	os.Stderr = nil
	log.SetOutput(io.Discard)
	Level = LEVEL_INFO
}
