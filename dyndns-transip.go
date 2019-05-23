package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"encoding/json"

	"github.com/transip/gotransip"
	"github.com/transip/gotransip/domain"
	"gopkg.in/yaml.v2"
)

type conf struct {
	Domain         string `yaml:"domain"`
	Server         string `yaml:"server"`
	PrivateKeyPath string `yaml:"privatekeypath"`
	AccountName    string `yaml:"accountname"`
	TimeToLive     int64  `yaml:"timetolive"`
}

type ipify struct {
	IP string `json:"ip"`
}

func (c *conf) getConf() *conf {

	yamlFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	fmt.Printf("Account: %s, Domain: %s, Server: %s\n", c.AccountName, c.Domain, c.Server)

	return c
}

func getMyIP() string {
	var ipifyresult ipify
	response, err := http.Get("https://api.ipify.org?format=json")
	if err != nil {
		panic(err.Error())
	}
	data, _ := ioutil.ReadAll(response.Body)
	json.Unmarshal([]byte(data), &ipifyresult)
	fmt.Printf("My IP Address: %s\n", ipifyresult.IP)

	return ipifyresult.IP
}

func getTransIPClient(accountName string, privateKeyPath string) gotransip.SOAPClient {
	// create new TransIP API SOAP client
	c, err := gotransip.NewSOAPClient(gotransip.ClientConfig{
		AccountName:    accountName,
		PrivateKeyPath: privateKeyPath})
	if err != nil {
		panic(err.Error())
	}

	return c
}

func getDomain(transipAPI gotransip.SOAPClient, domainName string) domain.Domain {

	// Get domain info from Transip
	dom, err := domain.GetInfo(transipAPI, domainName)
	if err != nil {
		panic(err.Error())
	}

	// print info for each DNS Entry
	// fmt.Print("DNS Entries:\n")
	// for _, dnsEntry := range dom.DNSEntries {
	// 	fmt.Printf("Name: %s, TTL: %d, Type: %s, Content: %s \n", dnsEntry.Name, dnsEntry.TTL, dnsEntry.Type, dnsEntry.Content)
	// }

	return dom
}

func isCorrectDNSValue(pDNSEntry *domain.DNSEntry, ip string, cfg conf) bool {
	return (*pDNSEntry).Content == ip && (*pDNSEntry).TTL == cfg.TimeToLive
}

func findDNSEntryForServer(dnsEntries domain.DNSEntries, serverName string) *domain.DNSEntry {
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
	var transipAPI gotransip.SOAPClient

	cfg.getConf()

	// create soap client for Transip API
	transipAPI = getTransIPClient(cfg.AccountName, cfg.PrivateKeyPath)
	myIP = getMyIP()
	dom := getDomain(transipAPI, cfg.Domain)
	pServerDNSEntry := findDNSEntryForServer(dom.DNSEntries, cfg.Server)

	if pServerDNSEntry != nil {
		fmt.Printf("Found server %s in DNS record\n", cfg.Server)
		if !isCorrectDNSValue(pServerDNSEntry, myIP, cfg) {
			(*pServerDNSEntry).Content = myIP
			(*pServerDNSEntry).TTL = cfg.TimeToLive

			err := domain.SetDNSEntries(transipAPI, cfg.Domain, dom.DNSEntries)
			if err != nil {
				panic(err.Error())
			}
			fmt.Printf("Server %s.%s has now ip adress %s with TTL %d \n", cfg.Server, cfg.Domain, myIP, cfg.TimeToLive)

		} else {
			fmt.Printf("Value of DNS is correct with ip %s, nothing will be changed\n", myIP)
		}
	} else {
		fmt.Printf("Server %s not found in DNS record\n", cfg.Server)
		entry := domain.DNSEntry{Name: cfg.Server, TTL: cfg.TimeToLive, Type: domain.DNSEntryTypeA, Content: myIP}
		dom.DNSEntries = append(dom.DNSEntries, entry)
		err := domain.SetDNSEntries(transipAPI, cfg.Domain, dom.DNSEntries)
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("Server %s has been added to DNS record of %s with ip adress %s \n", cfg.Server, cfg.Domain, myIP)
	}

}
