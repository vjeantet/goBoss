package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/vjeantet/portmap"
	"github.com/vjeantet/portmap/gateway"
)

func downloadFile(c *gin.Context) {
	token := c.Param("token")

	if token == conf.Token {
		file := conf.Files[0]
		if _, err := os.Stat(file.FilePath); err != nil {
			c.Header("Content-Type", "text/html")
			c.String(404, "<h1>%s</h1>%s", "Fichier introuvable", err)
			fmt.Printf("error - %s - %s\n", c.Request.RemoteAddr, err)
		} else {
			c.Header("Content-Disposition", "attachment; filename=\""+file.FileBaseName+"\"")
			c.File(file.FilePath)
			conf.DownloadCounter++
			fmt.Printf("%d - shared - %s - %s\n", conf.DownloadCounter, c.Request.RemoteAddr, file.FileBaseName)

		}
	} else {
		c.String(403, "Fichier inconnu")
		fmt.Printf("error - %s - unknow token '%s' \n", c.Request.RemoteAddr, token)
	}
}

type Server struct {
	ln net.Listener
	pm portmap.Mapping
}

func NewServer() *Server {
	return &Server{}
}

func MaxAllowed(n int) gin.HandlerFunc {
	return func(c *gin.Context) {
		// before request
		c.Next()
		// after request
		if conf.DownloadCounter >= conf.DownloadLimit {
			close(conf.Done)
		}
	}
}

func (s *Server) hasGateway() bool {
	if ips, err := gateway.GetIPs(); err == nil {
		return len(ips) > 0
	}
	return false
}

func (s *Server) Serve(wan bool) {
	// Configure Web Server
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = nil
	r := gin.Default()
	r.Use(MaxAllowed(conf.DownloadLimit))
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.GET("/g/:token", downloadFile)
	s.ln, _ = net.Listen("tcp", ":0")
	_, conf.LocalPort, _ = net.SplitHostPort(s.ln.Addr().String())

	go http.Serve(s.ln, r)

	if wan == true {
		conf.WanIp, conf.WanPort = s.ExposeWan(conf.LocalPort, 15)
	}
}

func (s *Server) Stop() {
	s.ln.Close()
	if s.pm != nil {
		fmt.Println("Suppression du paramétrage de la box...")
		s.pm.Delete()
		time.Sleep(4 * time.Second)
	}
}

func (s *Server) ExposeWan(localport string, lifetime int) (string, string) {
	// Expose link on internet
	portInt, _ := strconv.Atoi(localport)
	s.pm, _ = portmap.New(portmap.Config{
		Protocol:     portmap.TCP,
		Name:         "goBoss-" + conf.Token,
		InternalPort: uint16(portInt),
		ExternalPort: 0,
		Lifetime:     time.Duration(lifetime) * time.Minute,
	})
	fmt.Printf("Recherche de l'adresse de la box...\n\n")
	<-s.pm.NotifyChan()
	s.pm.StopBroadcast()
	externalAddr := strings.Split(s.pm.ExternalAddr(), ":")
	return externalAddr[0], externalAddr[1]
}
