package main

import (
	"fmt"
	"net"
)

func dig(host string) {
	ips, err := net.LookupIP(host)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	for _, ip := range ips {
		fmt.Printf("%v:%v\n", host, ip)
	}
}

func main() {
	dig("sftp01.neogov.com")
}
