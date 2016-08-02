package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"
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

	LocalPort     string
	LocalIP       string
	LocalHostname string
	LocalDomain   string
	WanPort       string
	WanIp         string
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
	host, ipv4, domain := getCurrentHostNameAndIPV4()

	return &Config{
		Files:         []*File{},
		Token:         token,
		LocalIP:       ipv4,
		Done:          make(chan bool),
		DownloadLimit: 1,
		LocalHostname: host,
		LocalDomain:   domain,
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

	if c.LocalDomain != "" {
		return fmt.Sprintf("http://%s.%s:%s/g/%s", c.LocalHostname, strings.ToLower(c.LocalDomain), c.LocalPort, c.Token)
	}

	return fmt.Sprintf("http://%s:%s/g/%s", c.LocalIP, c.LocalPort, c.Token)

}

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
