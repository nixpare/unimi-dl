package main

import (
	"net/http"
	"net/http/cookiejar"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/publicsuffix"
)

type UnimiDL struct {
	Lectures []*Lecture
	DLM      *DLManager
	pageURL  string
	Client 	 *http.Client
	rawPage  string
}

func NewUnimiDL(pageURL string) (*UnimiDL, error) {
	u := &UnimiDL{
		Lectures:  make([]*Lecture, 0),
		pageURL:   pageURL,
		Client: new(http.Client),
	}

	u.DLM = NewDLManager(u.Client)

	var err error
	u.Client.Jar, err = cookiejar.New(&cookiejar.Options {
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
	err := performLogin(u.Client)
	if err != nil {
		return err
	}

	u.rawPage, err = getPage(u.Client, u.pageURL)
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
		lectures = append(lectures, newLecture(u.pageURL, s, u.DLM))
	})

	return lectures, nil
}
