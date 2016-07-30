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
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Bowery/prompt"
	"github.com/atotto/clipboard"
	"github.com/gin-gonic/gin"
	"github.com/skratchdot/open-golang/open"
)

type Configuration struct {
	Token            string
	FilePath         string
	FileBaseName     string
	Password         string
	DelaisMinutes    string
	DelaisMinutesInt int
	EnvoyerParMail   bool
	// List            string   `form:"list" choices:"Choice 1, Choice 2"`
	// Checkbox        []string `form:"checkbox" choices:"Choice 1, Choice 2"`
}

func main() {
	if len(os.Args) < 2 {
		log.Println("donner un fichier svp")
		os.Exit(1)
	}
	argsWithoutProg := os.Args[1:]
	cheminFichier := strings.Join(argsWithoutProg, ", ")

	var conf Configuration

	conf.DelaisMinutes, _ = prompt.Basic("[ Desactiver le partage apres combien de minutes (10 par defaut) ? : ", false)
	conf.EnvoyerParMail, _ = prompt.Ask("[ Envoyer le lien par email ? : ")

	// ---

	conf.Token = GetMD5Hash(cheminFichier)
	conf.FileBaseName = filepath.Base(cheminFichier)
	conf.FilePath = filepath.Clean(cheminFichier)

	var err error
	if conf.DelaisMinutesInt, err = strconv.Atoi(conf.DelaisMinutes); err != nil {
		conf.DelaisMinutes = "10"
		conf.DelaisMinutesInt = 10
	}

	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = nil
	r := gin.Default()

	r.GET("/g/:token", func(c *gin.Context) {
		token := c.Param("token")

		if token == conf.Token {
			if _, err := os.Stat(conf.FilePath); err != nil {
				c.Header("Content-Type", "text/html")
				c.String(404, "<h1>%s</h1>%s", "Fichier introuvable", err)
				fmt.Printf("-- Erreur : %s\n", err)
			} else {
				c.Header("Content-Disposition", "attachment; filename=\""+conf.FileBaseName+"\"")
				c.File(conf.FilePath)
				fmt.Printf("-- Récupation OK du fichier %s par %s\n", conf.FileBaseName, c.Request.RemoteAddr)
			}
		} else {
			c.String(403, "Fichier inconnu")
			fmt.Printf("-- Erreur : token inconnu : '%s'\n", token)
		}
	})

	ln, _ := net.Listen("tcp", ":0")
	_, port, _ := net.SplitHostPort(ln.Addr().String())
	// fmt.Println("Listening on port", port)

	host, ipv4 := getCurrentHostNameAndIPV4()
	var lienHost string
	var lienIP string

	lienIP = fmt.Sprintf("http://%s:%s/g/%s", ipv4, port, conf.Token)
	if host == "" {
		lienHost = fmt.Sprintf("http://%s:%s/g/%s", host, port, conf.Token)
	}

	fmt.Printf("Mise à disposition du fichier %s\n\n", conf.FileBaseName)
	fmt.Printf("\t%s\n\n", lienHost)
	fmt.Printf("\t%s\n\n", lienIP)

	if conf.EnvoyerParMail {
		open.Run(fmt.Sprintf("mailto:?subject=Fichier pour vous&Body=%s", lienIP))
	}
	go http.Serve(ln, r)

	DiedAt := time.Now().Add(time.Duration(conf.DelaisMinutesInt) * time.Minute)

	fmt.Printf("\tLe partage se terminera à %s, dans %d minutes\n\n", DiedAt.Format("15:04:05"), conf.DelaisMinutesInt)

	if err := clipboard.WriteAll(lienIP); err == nil {
		fmt.Printf("\tLe lien 'http://..' a été copié dans le presse papier\n\n")
	}

	time.Sleep(time.Minute * time.Duration(conf.DelaisMinutesInt))
	ln.Close()

	fmt.Print("\a")
	fmt.Println("Partage terminé, vous pouvez fermer cette fenetre...")
	done := make(chan bool)
	<-done

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
