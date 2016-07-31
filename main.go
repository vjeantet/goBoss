package main

//go:generate goversioninfo -icon=win.ico

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/Bowery/prompt"
	"github.com/atotto/clipboard"
	"github.com/skratchdot/open-golang/open"
)

var conf *Config

func main() {
	var err error

	if len(os.Args) < 2 {
		log.Println("donner un fichier svp")
		os.Exit(1)
	}
	cheminsFichier := os.Args[1:]

	conf = NewConfig()
	for _, f := range cheminsFichier {
		conf.AddFilePath(f)
	}

	// Configuration
	// Ask for configuration
	conf.Internet, _ = prompt.Ask("[ Etes vous derriere une box sur internet ?")
	conf.DelaisMinutes, _ = prompt.Basic("[ Desactiver le partage au bout de combien de minutes (10 par defaut) ? : ", false)
	conf.EnvoyerParMail, _ = prompt.Ask("[ Envoyer le lien par email ? : ")

	if conf.DelaisMinutesInt, err = strconv.Atoi(conf.DelaisMinutes); err != nil {
		conf.DelaisMinutes = "10"
		conf.DelaisMinutesInt = 10
	}

	// Timer
	DiedAt := time.Now().Add(time.Duration(conf.DelaisMinutesInt) * time.Minute)
	timesUpChan := time.NewTicker(time.Minute * time.Duration(conf.DelaisMinutesInt)).C

	// WebServer
	server := NewServer()
	server.Serve(conf.Internet)

	fmt.Printf("Mise à disposition du fichier %s\n\n", conf.Files.String())
	fmt.Printf("\t%s\n\n", conf.Link())
	fmt.Printf("\tLe partage se terminera à %s, dans %d minutes\n\n", DiedAt.Format("15:04:05"), conf.DelaisMinutesInt)

	// Send link to clipboard
	if err := clipboard.WriteAll(conf.Link()); err == nil {
		fmt.Printf("\tLe lien a été copié dans le presse papier\n\n")
	}

	// Prepare email
	if conf.EnvoyerParMail {
		open.Run(fmt.Sprintf("mailto:?subject=Fichier pour vous&Body=%s", conf.Link()))
	}

	// Wait for signal CTRL+C for send a stop event to all AgentProcessor
	// When CTRL+C, SIGINT and SIGTERM signal occurs
	// Then stop server gracefully
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-timesUpChan:
		fmt.Print("\a")
		fmt.Println("Partage terminé...")
		close(ch)
	case <-ch:
		close(ch)
	}

	server.Stop()

}
