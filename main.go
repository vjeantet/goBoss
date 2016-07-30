package main

//go:generate goversioninfo -icon=win.ico

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Bowery/prompt"
	"github.com/atotto/clipboard"
	"github.com/gin-gonic/contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/skratchdot/open-golang/open"
	"github.com/vjeantet/portmap"
)

type Configuration struct {
	Token            string
	FilePath         string
	FileBaseName     string
	Password         string
	DelaisMinutes    string
	DelaisMinutesInt int
	EnvoyerParMail   bool
	Internet         bool
	// List            string   `form:"list" choices:"Choice 1, Choice 2"`
	// Checkbox        []string `form:"checkbox" choices:"Choice 1, Choice 2"`
}

var conf Configuration

func main() {
	var err error

	if len(os.Args) < 2 {
		log.Println("donner un fichier svp")
		os.Exit(1)
	}
	argsWithoutProg := os.Args[1:]
	cheminFichier := strings.Join(argsWithoutProg, ", ")

	// Configuration
	conf.Token = GetMD5Hash(cheminFichier + time.Now().String())
	conf.FileBaseName = filepath.Base(cheminFichier)
	conf.FilePath = filepath.Clean(cheminFichier)
	// Ask for configuration
	conf.Internet, _ = prompt.Ask("[ Etes vous derriere une box sur internet ?")
	conf.DelaisMinutes, _ = prompt.Basic("[ Desactiver le partage au bout de combien de minutes (10 par defaut) ? : ", false)
	conf.EnvoyerParMail, _ = prompt.Ask("[ Envoyer le lien par email ? : ")

	if conf.DelaisMinutesInt, err = strconv.Atoi(conf.DelaisMinutes); err != nil {
		conf.DelaisMinutes = "10"
		conf.DelaisMinutesInt = 10
	}

	// Configure Web Server
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = nil
	r := gin.Default()
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.GET("/g/:token", downloadFile)
	ln, _ := net.Listen("tcp", ":0")
	_, port, _ := net.SplitHostPort(ln.Addr().String())
	go http.Serve(ln, r)

	// Timer
	DiedAt := time.Now().Add(time.Duration(conf.DelaisMinutesInt) * time.Minute)
	timesUpChan := time.NewTicker(time.Minute * time.Duration(conf.DelaisMinutesInt)).C

	// Format Link
	var lienIP string
	_, ipv4 := getCurrentHostNameAndIPV4()
	lienIP = fmt.Sprintf("http://%s:%s/g/%s", ipv4, port, conf.Token)

	// Expose link on internet
	var m portmap.Mapping
	if conf.Internet == true {
		portInt, _ := strconv.Atoi(port)
		m, _ = portmap.New(portmap.Config{
			Protocol:     portmap.TCP,
			Name:         "goBoss-" + conf.Token,
			InternalPort: uint16(portInt),
			ExternalPort: 0,
			Lifetime:     time.Duration(conf.DelaisMinutesInt) * time.Minute,
		})
		fmt.Printf("Recherche de l'adresse de la box...\n\n")
		<-m.NotifyChan()
		m.StopBroadcast()
		lienIP = fmt.Sprintf("http://%s/g/%s", m.ExternalAddr(), conf.Token)
	}

	fmt.Printf("Mise à disposition du fichier %s\n\n", conf.FileBaseName)
	fmt.Printf("\t%s\n\n", lienIP)
	fmt.Printf("\tLe partage se terminera à %s, dans %d minutes\n\n", DiedAt.Format("15:04:05"), conf.DelaisMinutesInt)

	// Send link to clipboard
	if err := clipboard.WriteAll(lienIP); err == nil {
		fmt.Printf("\tLe lien a été copié dans le presse papier\n\n")
	}

	// Prepare email
	if conf.EnvoyerParMail {
		open.Run(fmt.Sprintf("mailto:?subject=Fichier pour vous&Body=%s", lienIP))
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

	if conf.Internet == true {
		m.Delete()
		fmt.Println("Suppression du paramétrage de la box...")
		time.Sleep(time.Second * 3)
	}

	ln.Close()

}

func getCurrentHostNameAndIPV4() (string, string) {
	name, err := os.Hostname()
	if err != nil {
		fmt.Printf("Oops: %v\n", err)
		name = ""
	}

	ifaces, err := net.Interfaces()
	_ = err
	// handle err
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		_ = err
		// handle err
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if !ip.IsLoopback() && ip.To4() != nil {
				return name, ip.String()
			}
		}
	}

	return name, ""
}

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
