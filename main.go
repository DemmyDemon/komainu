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

	cwd_override := os.Getenv("CWD_OVERRIDE")
	if cwd_override == "" {
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
		log.Println("CWD_OVERRIDE in effect: " + cwd_override)
		err := os.Chdir(cwd_override)
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
	log.Println("Preparing to connect to Discord")

	state := bot.Connect(&cfg)
	defer state.Close()

	WaitForInterrupt()

}

func WaitForInterrupt() {
	// Thanks to various Discord Gophers for this very simple stuff.
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-signalCh
}
