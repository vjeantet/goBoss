package main

//go:generate goversioninfo -icon=win.ico

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/Bowery/prompt"
	"github.com/atotto/clipboard"
	"github.com/skratchdot/open-golang/open"
)

var conf *Config

func main() {
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

	// WebServer
	server := NewServer()
	server.Serve(conf.Internet)

	fmt.Printf("Mise à disposition du fichier %s\n\n", conf.Files.String())
	fmt.Printf("\t%s\n\n", conf.Link())
	fmt.Printf("\tLe partage s'arretera des que le fichier sera téléchargé\n\n")

	// Send link to clipboard
	if err := clipboard.WriteAll(conf.Link()); err == nil {
		fmt.Printf("\tLe lien a été copié dans le presse papier\n\n")
	}

	// Send by mail ?
	go func() {
		if ok, _ := prompt.Ask("[ Envoyer le lien par email ? : "); ok {
			open.Run(fmt.Sprintf("mailto:?subject=Fichier pour vous&Body=%s", url.QueryEscape(conf.Link())))
		}
	}()

	// Wait for signal CTRL+C for send a stop event to all AgentProcessor
	// When CTRL+C, SIGINT and SIGTERM signal occurs
	// Then stop server gracefully
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ch:
		close(ch)
	}

	server.Stop()

}
