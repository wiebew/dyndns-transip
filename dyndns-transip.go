package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"log"

	"github.com/transip/gotransip/v6"
	"github.com/transip/gotransip/v6/domain"
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

func (c *conf) getConf(path *string) *conf {

	yamlFile, err := ioutil.ReadFile(*path)
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
		log.Fatal(err.Error())
	}
	data, _ := ioutil.ReadAll(response.Body)
	json.Unmarshal([]byte(data), &ipifyresult)
	log.Printf("My IP Address: %s\n", ipifyresult.IP)

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

	var configpath = flag.String("config", "./config.yaml", "path to config.yml file")
	flag.Parse()

	cfg.getConf(configpath)
	log.SetOutput(os.Stdout)
	log.Printf("Check on Account: %s, Domain: %s, Server: %s\n", cfg.AccountName, cfg.Domain, cfg.Server)

	// create soap client for Transip API
	transipAPI, err := gotransip.NewClient(gotransip.ClientConfiguration{
		AccountName:    cfg.AccountName,
		PrivateKeyPath: cfg.PrivateKeyPath})
	if err != nil {
		log.Fatal(err.Error())
	}

	myIP = getMyIP()

	myDomain := domain.Repository{Client: transipAPI}

	dnsEntries, err = myDomain.GetDNSEntries(cfg.Domain)
	if err != nil {
		log.Fatal(err.Error())
	}

	pServerDNSEntry := findDNSEntryForServer(dnsEntries, cfg.Server)

	if pServerDNSEntry != nil {
		log.Printf("Found server %s in DNS record\n", cfg.Server)
		if !isCorrectDNSValue(pServerDNSEntry, myIP, cfg) {
			(*pServerDNSEntry).Content = myIP
			(*pServerDNSEntry).Expire = cfg.TimeToLive

			err := myDomain.ReplaceDNSEntries(cfg.Domain, dnsEntries)
			if err != nil {
				panic(err.Error())
			}
			log.Printf("Server %s.%s now has ip address %s with TTL %d \n", cfg.Server, cfg.Domain, myIP, cfg.TimeToLive)

		} else {
			log.Printf("DNS entry is correct with IP %s, nothing will be changed\n", myIP)
		}
	} else {
		log.Printf("Server %s not found in DNS record\n", cfg.Server)
		entry := domain.DNSEntry{Name: cfg.Server, Expire: cfg.TimeToLive, Type: "A", Content: myIP}
		dnsEntries = append(dnsEntries, entry)
		err := myDomain.ReplaceDNSEntries(cfg.Domain, dnsEntries)
		if err != nil {
			panic(err.Error())
		}
		log.Printf("Server %s has now been added to DNS record of %s with IP address %s with TTL %d \n", cfg.Server, cfg.Domain, myIP, cfg.TimeToLive)
	}
}
