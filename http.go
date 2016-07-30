package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
)

func downloadFile(c *gin.Context) {
	token := c.Param("token")

	if token == conf.Token {
		if _, err := os.Stat(conf.FilePath); err != nil {
			c.Header("Content-Type", "text/html")
			c.String(404, "<h1>%s</h1>%s", "Fichier introuvable", err)
			fmt.Printf("error - %s - %s\n", c.Request.RemoteAddr, err)
		} else {
			c.Header("Content-Disposition", "attachment; filename=\""+conf.FileBaseName+"\"")
			c.File(conf.FilePath)
			fmt.Printf("shared - %s - %s\n", c.Request.RemoteAddr, conf.FileBaseName)
		}
	} else {
		c.String(403, "Fichier inconnu")
		fmt.Printf("error - %s - unknow token '%s' \n", c.Request.RemoteAddr, token)
	}
}
