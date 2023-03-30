package main

import (
	"net/http"
	"net/http/cookiejar"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/publicsuffix"
)

type UnimiDL struct {
	Lectures  []*Lecture
	dlManager *dlManager
	pageURL   string
	client 	  *http.Client
	rawPage   string
}

func NewUnimiDL(pageURL string) (*UnimiDL, error) {
	u := &UnimiDL{
		Lectures:  make([]*Lecture, 0),
		pageURL:   pageURL,
		client: new(http.Client),
	}

	u.dlManager = newDLManager(u.client)

	var err error
	u.client.Jar, err = cookiejar.New(&cookiejar.Options {
		PublicSuffixList: publicsuffix.List,
	})
	return u, err
}

func (u *UnimiDL) GetAllLectures() error {
	err := u.getPageOnline()
	if err != nil {
		return err
	}

	u.Lectures, err = u.findAllLectures(u.rawPage)
	return err
}

func (u *UnimiDL) getPageOnline() error {
	err := performLogin(u.client)
	if err != nil {
		return err
	}

	u.rawPage, err = getPage(u.client, u.pageURL)
	return err
}

func (u *UnimiDL) findAllLectures(page string) ([]*Lecture, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(page))
	if err != nil {
		return nil, err
	}

	lectures := make([]*Lecture, 0)

	sel := doc.Find(LECTURE_QUERY)
	sel.Each(func(i int, s *goquery.Selection) {
		lectures = append(lectures, newLecture(u.pageURL, s, u.dlManager))
	})

	return lectures, nil
}
