package main

import (
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	LECTURE_QUERY = "#threadList > tr.sticky"
	LECTURE_TITLE_QUERY = "h2.arielTitle.arielStick span"
)

const (
	VIDEO_WRAPPER_QUERY = "div.lecturecVideo"
	VIDEO_QUERY = "video.lecturec"
)

type Lecture struct {
	selection 	*goquery.Selection
	raw 		string
	title 		string
	videos 		[]Video
}

type Video struct {
	uploadTime 	time.Time
	manifestURL string
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
		
	titleSpans := s.Find(LECTURE_TITLE_QUERY)
	titleSplit := strings.Split(
		strings.ReplaceAll(titleSpans.Last().Text(), "\t", " ",),
		" ",
	)
	
	title := make([]string, 0)
	for _, s := range titleSplit {
		if s == "" {
			continue
		}
		title = append(title, strings.TrimSpace(s))
	}

	return &Lecture {
		raw: raw,
		title: strings.Join(title, " "),
	}
}
