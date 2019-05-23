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
	Domain string `yaml:"domain"`
	Server string `yaml:"server"`
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

	return ipifyresult.IP
}

func getSoapClient(accountName string, privateKeyPath string) SOAPClient {
	// create new TransIP API SOAP client
	c, err := gotransip.NewSOAPClient(gotransip.ClientConfig{
		AccountName:    "",
		PrivateKeyPath: ""})
	if err != nil {
		panic(err.Error())
	}

	return c
}

func main() {
	var cfg conf
	var myIP string
	var transIP SOAPClient

	cfg.getConf()
	domainName := cfg.Domain
	serverName := cfg.Server

	fmt.Printf("Domain: %s, Server: %s\n", domainName, serverName)

	// fetch ipaddress
	myIP = getMyIP()
	fmt.Printf("My IP Address: %s\n", myIP)

	// get list of domains
	domain, err := domain.GetInfo(c, domainName)
	if err != nil {
		panic(err.Error())
	}

	// print info for each DNS Entry
	fmt.Print("DNS Entries:\n")
	for _, v := range domain.DNSEntries {
		fmt.Printf("Name: %s, TTL: %d, Type: %s, Content: %s \n", v.Name, v.TTL, v.Type, v.Content)
	}

	found := false
	for _, v := range domain.DNSEntries {
		if v.Name == serverName {
			found = true
			fmt.Printf("Found server %s in DNS record\n", serverName)
		}
	}
	if !found {
		fmt.Printf("Server %s not found in DNS record\n", serverName)
	}

}
