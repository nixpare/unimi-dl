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
	Title 		string
	Message 	string
	Videos 		[]Video
	Attachments []Attachment
	selection 	*goquery.Selection
	raw 		string
	pageURL 	string
}

func FindAllLectures(page string, pageURL string) ([]*Lecture, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(page))
	if err != nil {
		return nil, err
	}

	lectures := make([]*Lecture, 0)

	sel := doc.Find(LECTURE_QUERY)
	sel.Each(func(i int, s *goquery.Selection) {
		lectures = append(lectures, NewLecture(pageURL, s))
	})

	return lectures, nil
}

func NewLecture(pageURL string, s *goquery.Selection) *Lecture {
	raw, _ := s.Html()
	titleSel := s.Find(LECTURE_TITLE_QUERY).Last()

	lecture := &Lecture {
		Title: ExtractTextFromHTML(titleSel),
		Videos: make([]Video, 0),
		Attachments: make([]Attachment, 0),
		selection: s,
		raw: raw,
		pageURL: pageURL,
	}

	lecture.findMessage()
	lecture.findVideos()
	lecture.findAttachments()

	return lecture
}

func (l *Lecture) findMessage() {
	sel := l.selection.Find(LECTURE_MESSAGE_QUERY)
	l.Message = ExtractTextFromHTML(sel)
}

func (l *Lecture) findVideos() {
	l.selection.Find(VIDEO_QUERY).Each(func(i int, s *goquery.Selection) {
		videoName := fmt.Sprintf("%s_%d", l.Title, i+1)
		l.Videos = append(l.Videos, NewVideo(videoName, s))
	})
}

func (l *Lecture) findAttachments() {
	l.selection.Find(ATTACHMENT_QUERY).Each(func(i int, s *goquery.Selection) {
		l.Attachments = append(l.Attachments, NewAttachment(l.pageURL, s))
	})
}

func (l Lecture) String() string {
	res := fmt.Sprintf(
		"Lecture Title: %s\nMessage: %s",
		l.Title,
		IndentMultilineString(l.Message, 9),
	)

	res += "\nVideos:"
	if len(l.Videos) == 0 {
		return res + " None"
	}

	for _, v := range l.Videos {
		res += fmt.Sprintf("\n  - %s", IndentMultilineString(v, 5))
	}

	res += "\nAttachments:"
	if len(l.Attachments) == 0 {
		return res + " None"
	}

	for _, a := range l.Attachments {
		res += fmt.Sprintf("\n  - %s", IndentMultilineString(a, 12))
	}
	return res
}
