package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"
)

type File struct {
	FilePath     string
	FileBaseName string
}

type FileList []*File

type Config struct {
	Token           string
	Password        string
	DownloadLimit   int
	DownloadCounter int
	Internet        bool

	Done chan bool

	Files FileList

	LocalPort string
	LocalIP   string
	WanPort   string
	WanIp     string
}

func (c *Config) AddFilePath(afile string) {
	file := &File{
		FilePath:     filepath.Clean(afile),
		FileBaseName: filepath.Base(afile),
	}
	c.Files = append(c.Files, file)
}

func NewConfig() *Config {
	token := GetMD5Hash(time.Now().String())
	_, ipv4 := getCurrentHostNameAndIPV4()
	return &Config{
		Files:         []*File{},
		Token:         token,
		LocalIP:       ipv4,
		Done:          make(chan bool),
		DownloadLimit: 1,
	}
}

func (f *FileList) String() string {
	s := ""
	for _, v := range *f {
		s = v.FileBaseName
	}
	return s
}

func (c *Config) Link() string {
	if c.WanPort != "" {
		return fmt.Sprintf("http://%s:%s/g/%s", c.WanIp, c.WanPort, c.Token)
	}
	return fmt.Sprintf("http://%s:%s/g/%s", c.LocalIP, c.LocalPort, c.Token)
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
			if !ip.IsLoopback() && ip.To4() != nil && !ip.IsLinkLocalUnicast() {
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
