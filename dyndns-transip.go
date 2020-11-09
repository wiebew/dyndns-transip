package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/google/logger"

	"github.com/transip/gotransip"
	"github.com/transip/gotransip/domain"
	"gopkg.in/yaml.v2"
)

type conf struct {
	Domain         string `yaml:"domain"`
	Server         string `yaml:"server"`
	PrivateKeyPath string `yaml:"privatekeypath"`
	AccountName    string `yaml:"accountname"`
	TimeToLive     int    `yaml:"timetolive"`
	LogFile        string `yaml:"logfile"`
}

type ipify struct {
	IP string `json:"ip"`
}

func (c *conf) getConf() *conf {

	yamlFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		fmt.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		fmt.Printf("Unmarshal: %v", err)
	}

	return c
}

func getMyIP() string {
	var ipifyresult ipify
	response, err := http.Get("https://api.ipify.org?format=json")
	if err != nil {
		logger.Fatal(err.Error())
	}
	data, _ := ioutil.ReadAll(response.Body)
	json.Unmarshal([]byte(data), &ipifyresult)
	logger.Infof("My IP Address: %s\n", ipifyresult.IP)

	return ipifyresult.IP
}

func isCorrectDNSValue(pDNSEntry *domain.DNSEntry, ip string, cfg conf) bool {
	return (*pDNSEntry).Content == ip && (*pDNSEntry).Expire == cfg.TimeToLive
}

func findDNSEntryForServer(dnsEntries []domain.DNSEntry, serverName string) *domain.DNSEntry {
	// search for servername in the dns entries
	var pServerDNSEntry *domain.DNSEntry

	// use idx to get correct reference in dnsEntries array
	for idx := range dnsEntries {
		dnsEntry := dnsEntries[idx]
		if dnsEntry.Name == serverName && dnsEntry.Type == "A" {
			pServerDNSEntry = &dnsEntries[idx]
		}
	}
	return pServerDNSEntry
}

func updateDNSRecord() {

}

func main() {
	var cfg conf
	var myIP string
	var dnsEntries []domain.DNSEntry

	cfg.getConf()
	flag.Parse()
	var verbose = flag.Bool("verbose", false, "print info level logs to stdout")

	lf, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		logger.Fatalf("Failed to open log file: %v", err)
	}
	defer lf.Close()

	defer logger.Init("DynDNSLogger", *verbose, true, lf).Close()
	logger.Infof("*** Starting dyndns-transip check on Account: %s, Domain: %s, Server: %s\n", cfg.AccountName, cfg.Domain, cfg.Server)

	// create soap client for Transip API
	transipAPI, err := gotransip.NewClient(gotransip.ClientConfiguration{
		AccountName:    cfg.AccountName,
		PrivateKeyPath: cfg.PrivateKeyPath})
	if err != nil {
		logger.Fatal(err.Error())
	}

	myIP = getMyIP()

	myDomain := domain.Repository{Client: transipAPI}

	dnsEntries, err = myDomain.GetDNSEntries(cfg.Domain)
	if err != nil {
		logger.Fatal(err.Error())
	}

	pServerDNSEntry := findDNSEntryForServer(dnsEntries, cfg.Server)

	if pServerDNSEntry != nil {
		logger.Infof("Found server %s in DNS record\n", cfg.Server)
		if !isCorrectDNSValue(pServerDNSEntry, myIP, cfg) {
			(*pServerDNSEntry).Content = myIP
			(*pServerDNSEntry).Expire = cfg.TimeToLive

			err := myDomain.ReplaceDNSEntries(cfg.Domain, dnsEntries)
			if err != nil {
				panic(err.Error())
			}
			logger.Infof("Server %s.%s has now ip address %s with TTL %d \n", cfg.Server, cfg.Domain, myIP, cfg.TimeToLive)

		} else {
			logger.Infof("Value of DNS is correct with ip %s, nothing will be changed\n", myIP)
		}
	} else {
		logger.Infof("Server %s not found in DNS record\n", cfg.Server)
		entry := domain.DNSEntry{Name: cfg.Server, Expire: cfg.TimeToLive, Type: "A", Content: myIP}
		dnsEntries = append(dnsEntries, entry)
		err := myDomain.ReplaceDNSEntries(cfg.Domain, dnsEntries)
		if err != nil {
			panic(err.Error())
		}
		logger.Infof("Server %s has been added to DNS record of %s with ip address %s \n", cfg.Server, cfg.Domain, myIP)
	}

}
