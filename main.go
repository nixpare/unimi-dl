package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"

	"golang.org/x/net/publicsuffix"
)

const TEST_PAGE = "https://nbasilicoae2.ariel.ctu.unimi.it/v5/frm3/ThreadList.aspx?fc=qBg4sBrRnwcdhrbedslZntFd2HdJGwehSpagKzRGGL46du5ML7nAZ1F3iVRHQ0jk&roomid=227362"

var client *http.Client

func main() {
	defer func() {
		fmt.Println("\nPress Enter to exit")
		var b []byte
		fmt.Scanf("%s\n", &b)
	}()

	logF, _ := os.OpenFile("unimi-dl.log", os.O_APPEND | os.O_CREATE | os.O_WRONLY, 0777)
	log.SetOutput(logF)

	page, err := GetPageOnline(TEST_PAGE)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	lectures, err := FindAllLectures(page, TEST_PAGE)
	if err != nil {
		panic(err)
	}

	for _, l := range lectures {
		fmt.Println(l)
		fmt.Println()
	}

	att := lectures[0].Attachments[0]

	fmt.Printf("Downloading attachment %s with url <%s>\n", att.Name, att.URL)
	att.Download()
}

func GetPageOnline(pageURL string) (string, error) {
	client = new(http.Client)
	cookieJar, err := cookiejar.New(&cookiejar.Options {
		PublicSuffixList: publicsuffix.List,
	})
	if err != nil {
		log.Printf("HTTP client error: %v\n", err)
		return "", errors.New("error while initializing the HTTP client")
	}
	client.Jar = cookieJar

	err = performLogin(client)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	page, err := getPage(client, pageURL)
	if err != nil {
		return "", err
	}

	return page, nil
}
