package main

import (
	"komainu/bot"
	"komainu/storage"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

func main() {

	if os.Getenv("DEV_MODE") != "" {
		log.SetFlags(log.Lshortfile | log.Ltime)
	}

	cwdOverride := os.Getenv("CWD_OVERRIDE")
	if cwdOverride == "" {
		exec, err := os.Executable()
		if err != nil {
			log.Fatalln("Failed to get executable path:", err)
		}
		path := filepath.Dir(exec)
		err = os.Chdir(path)
		if err != nil {
			log.Fatalln("Failed to change active directory to where the executable lives:", err)
		}
		log.Printf("Spinning up Komainu in %s\n", path)
	} else {
		log.Println("CWD_OVERRIDE in effect: " + cwdOverride)
		err := os.Chdir(cwdOverride)
		if err != nil {
			log.Fatalln("Failed to change active directory:", err)
		}
	}

	cfg := storage.GetConfiguration()
	if cfg.Logfile != "" {
		log.Printf("Using %s for a log file\n", cfg.Logfile)
		if logfileHandle, err := os.OpenFile(cfg.Logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640); err == nil {
			log.SetOutput(logfileHandle)
			defer logfileHandle.Close()
		}
	} else {
		log.Println("No logfile specified: Will just output to STDOUT and hope for the best.")
	}

	kvs, err := storage.OpenBolt("data/bolt")
	if err != nil {
		log.Fatalln("Could not open KVS:", err)
	}
	defer kvs.Close()

	log.Println("Preparing to connect to Discord")

	state := bot.Connect(&cfg, kvs)
	defer state.Close()

	WaitForInterrupt()

}

// WaitForInterrupt blocks until a SIGINT, SIGTERM or another OS interrupt is received.
// "Pause until Ctrl+C", basically.
func WaitForInterrupt() {
	// Thanks to various Discord Gophers for this very simple stuff.
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGTERM, os.Interrupt)
	<-signalCh
}
