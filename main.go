package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"

	"golang.org/x/net/publicsuffix"
)

func main() {
	defer func() {
		fmt.Println("\nPress Enter to exit")
		var b []byte
		fmt.Scanf("%s\n", &b)
	}()

	logF, _ := os.OpenFile("unimi-dl.log", os.O_APPEND | os.O_CREATE | os.O_WRONLY, 0777)
	log.SetOutput(logF)

	client := new(http.Client)
	cookieJar, err := cookiejar.New(&cookiejar.Options {
		PublicSuffixList: publicsuffix.List,
	})
	if err != nil {
		fmt.Printf("Error while initializing the HTTP client")
		log.Printf("HTTP client error: %v\n", err)
		os.Exit(1)
	}
	client.Jar = cookieJar

	err = performLogin(client)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	page, err := getPage(
		client,
		"https://nbasilicoae2.ariel.ctu.unimi.it/v5/frm3/ThreadList.aspx?fc=qBg4sBrRnwcdhrbedslZntFd2HdJGwehSpagKzRGGL46du5ML7nAZ1F3iVRHQ0jk&roomid=227362",
	)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	lectures, err := FindAllLectures(page)
	if err != nil {
		panic(err)
	}

	for _, l := range lectures {
		fmt.Println(l)
		fmt.Println()
	}
}
