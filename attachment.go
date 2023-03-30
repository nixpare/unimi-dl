package main

import (
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const ATTACHMENT_QUERY = ".arielAttachmentBox tr a.filename"

type Attachment struct {
	Name 	  string
	URL 	  string
	dlManager *dlManager
}

func newAttachment(pageURL string, sel *goquery.Selection, dlManager *dlManager) Attachment {
	pageURL = strings.Split(pageURL, "?")[0]
	parsedURL, _ := url.ParseRequestURI(pageURL)
	href, _ := url.PathUnescape(parsedURL.JoinPath(sel.AttrOr("href", "")).String())

	return Attachment {
		Name: sel.Text(),
		URL: strings.ReplaceAll(href, "frm3/frm3/", "frm3/"),
		dlManager: dlManager,
	}
}

func (a Attachment) Download(prefix string) error {
	failErr := fmt.Errorf("an error has occurred while downloading attachment %s", a.Name)

	resp, err := a.dlManager.client.Get(a.URL)
	if err != nil || resp.StatusCode >= 400 {
		log.Printf("Error downloading attachment %s (Code %d): %v\n", a.Name, resp.StatusCode, err)
		return failErr
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response for attachment %s: %v\n", a.Name, err)
		return failErr
	}

	prefix = strings.ReplaceAll(prefix, "/", "-")
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	dirPath, _ := os.Getwd()
	filePath := dirPath + "/" + prefix + strings.ReplaceAll(a.Name, "/", "-")
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0777)
	if err != nil {
		log.Printf("Error creating file for attachment %s: %v\n", a.Name, err)
		return failErr
	}
	defer file.Close()

	file.Write(respBody)
	return nil
}

func (a Attachment) String() string {
	return a.Name
}
