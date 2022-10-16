package utils

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

var mux = &sync.Mutex{}
var verbose, help bool
var logFile string

const (
	ClientID = "X-Client-ID"
)

var LogWriter zerolog.Logger

func Init() {

	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	flag.StringVar(&logFile, "o", "", "output file path, if empty logs will print on standard output, example -o /tmp/log.txt")
	flag.BoolVar(&verbose, "v", false, "enables verbose logging")
	flag.BoolVar(&help, "help", false, "prints available flags")
	flag.Parse()

	if help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if logFile != "" {
		UpdateLogger(logFile)
	}

	if verbose {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	LogWriter = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006-01-02 03:04:05 PM"}).With().Timestamp().Logger()

	LogWriter.Info().Msgf("Number of cpu's are %d", runtime.NumCPU())
	PrintFlags()
}

func UpdateLogger(fpath string) {
	f, err := os.OpenFile(fpath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}
	LogWriter = LogWriter.Output(f)
}

func PrintFlags() {
	flag.VisitAll(func(f *flag.Flag) {
		LogWriter.Info().Msgf("key : %v, value : %v", f.Name, f.Value)
	})
}

func Incr(i *int) func() {

	mux.Lock()
	defer mux.Unlock()
	*i++

	return func() {
		mux.Lock()
		defer mux.Unlock()
		*i--
	}
}

func UpdateMemoryUsage(logger *zerolog.Logger, tot int) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	info := logger.Info()
	info.Int("Active Conn", tot)
	info.Str("Alloc", ByteCountIEC(uint(m.Alloc)))
	info.Str("TotalAlloc", ByteCountIEC(uint(m.TotalAlloc)))
	info.Str("Sys", ByteCountIEC(uint(m.Sys)))
	info.Int("NumGC", int(m.NumGC))
	info.Msg("stats")
}

func PrintCurrentActiveConnections(logger *zerolog.Logger, totalConnections *int) {

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		UpdateMemoryUsage(logger, *totalConnections)
	}

}

func ByteCountIEC(b uint) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}
