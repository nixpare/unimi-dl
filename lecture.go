package main

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	LECTURE_QUERY = "#threadList > tr.sticky"
	LECTURE_TITLE_QUERY = "h2.arielTitle.arielStick span"
	LECTURE_MESSAGE_QUERY = ".arielMessageBody > span.postbody"
)

type Lecture struct {
	selection 	*goquery.Selection
	raw 		string
	Title 		string
	Message 	string
	Videos 		[]Video
}

func FindAllLectures(page string) ([]*Lecture, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(page))
	if err != nil {
		return nil, err
	}

	lectures := make([]*Lecture, 0)

	sel := doc.Find(LECTURE_QUERY)
	sel.Each(func(i int, s *goquery.Selection) {
		lectures = append(lectures, NewLecture(s))
	})

	return lectures, nil
}

func NewLecture(s *goquery.Selection) *Lecture {
	raw, _ := s.Html()
	titleSel := s.Find(LECTURE_TITLE_QUERY).Last()

	lecture := &Lecture {
		raw: raw,
		Title: ExtractTextFromHTML(titleSel),
		selection: s,
		Videos: make([]Video, 0),
	}

	lecture.findMessage()
	lecture.findVideos()

	return lecture
}

func (l *Lecture) findMessage() {
	sel := l.selection.Find(LECTURE_MESSAGE_QUERY)
	l.Message = ExtractTextFromHTML(sel)
}

func (l *Lecture) findVideos() {
	l.selection.Find(VIDEO_QUERY).Each(func(i int, s *goquery.Selection) {
		l.Videos = append(l.Videos, NewVideo(s))
	})
}

func (l Lecture) String() string {
	res := fmt.Sprintf(
		"Lecture Title: %s\nMessage: %s\nVideos:",
		l.Title,
		l.Message,
	)
	if len(l.Videos) == 0 {
		return res + " None"
	}

	for _, v := range l.Videos {
		res += fmt.Sprintf("\n  -  %s", IndentMultilineString(v, 5))
	}
	return res
}
