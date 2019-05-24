# dyndns-transip

[![Go Report Card](https://goreportcard.com/badge/github.com/wiebew/dyndns-transip)](https://goreportcard.com/report/github.com/wiebew/dyndns-transip) [![Documentation]

Dyndns script that dynamically sets ip address in DNS records of a domain with internet provider [Transip](https://transip.nl).

This script is meant to let a server detect its own ip address on the internet using `ipify.org` and set that ip address in a transip DNS Entry using the `transip API`

Steps:

1. Activate API in the [Transip Portal](https://www.transip.nl/cp/account/api/), store the private key in your .ssh folder
2. Clone this project on your server
3. Copy the `config.yaml.example` to `config.yaml` and change the values to reflect your wishes
4. Make sure golang is installed on your machine, setup GOPATH and GOBIN variables
5. Do the following commands:

```bash
cd dyndns-transip
go get
go build
```

the result of this will be a binary `dyndns-transip` that you can call from a cronjob. Running this binary will update the ipaddresses opr TTL automatically when they differ from the desired situation.