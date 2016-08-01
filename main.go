package main

//go:generate goversioninfo -icon=win.ico

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/Bowery/prompt"
	"github.com/atotto/clipboard"
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

	// Ask for configuration
	max, _ := prompt.BasicDefault("Quitter apres combien de téléchargement ?", "1")
	conf.DownloadLimit, _ = strconv.Atoi(max)

	// WebServer
	server := NewServer()
	server.Serve(conf.Internet)

	fmt.Printf("Mise à disposition du fichier %s\n\n", conf.Files.String())
	fmt.Printf("\t%s\n\n", conf.Link())
	fmt.Printf("\tLe partage s'arretera des que le fichier sera téléchargé %d fois\n\n", conf.DownloadLimit)

	if err := clipboard.WriteAll(conf.Link()); err == nil {
		fmt.Printf("\tLe lien a été copié dans le presse papier\n\n")
	}

	// More Download limit ?
	go func(*Config) {
		if !server.hasGateway() {
			return
		}
		conf.Internet, _ = prompt.Ask("[ Partager sur internet (port mapping UPNP) ?")
		if conf.Internet == true {
			conf.WanIp, conf.WanPort = server.ExposeWan(conf.LocalPort, 15)
			fmt.Printf("\t%s\n\n", conf.Link())

			// Send link to clipboard
			if err := clipboard.WriteAll(conf.Link()); err == nil {
				fmt.Printf("\tLe lien a été copié dans le presse papier\n\n")
			}
		}
	}(conf)

	// Wait for signal CTRL+C for send a stop event to all AgentProcessor
	// When CTRL+C, SIGINT and SIGTERM signal occurs
	// Then stop server gracefully
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-conf.Done:
		close(ch)
	case <-ch:
		close(ch)
	}

	server.Stop()

}
