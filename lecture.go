package main

import (
	"fmt"
	"net/http"
	"os"

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
	Videos 		[]*Video
	Attachments []*Attachment
	selection 	*goquery.Selection
	raw 		string
	pageURL 	string
	dlManager 	*DLManager
}

func newLecture(pageURL string, s *goquery.Selection, dlManager *DLManager) *Lecture {
	raw, _ := s.Html()
	titleSel := s.Find(LECTURE_TITLE_QUERY).Last()

	lecture := &Lecture {
		Title: ExtractTextFromHTML(titleSel),
		Videos: make([]*Video, 0),
		Attachments: make([]*Attachment, 0),
		selection: s,
		raw: raw,
		pageURL: pageURL,
		dlManager: dlManager,
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
		l.Videos = append(l.Videos, newVideo(videoName, s))
	})
}

func (l *Lecture) findAttachments() {
	l.selection.Find(ATTACHMENT_QUERY).Each(func(i int, s *goquery.Selection) {
		l.Attachments = append(l.Attachments, newAttachment(l.pageURL, s))
	})
}

func (l *Lecture) Download() {
	info, err := os.Stat(l.Title)
	if err != nil || !info.IsDir() {
		os.Mkdir(l.Title, 0777) 		// Unhandled error
	}

	for _, a := range l.Attachments {
		fmt.Printf("Downloading attachment %s of lecture %s\n", a.Name, l.Title)

		el := l.dlManager.Download(
			a.Name,
			func(client *http.Client) error {
				return a.Download(client, l.Title)
			},
		)

		go func(a *Attachment) {
			err := el.Wait()
			if err == nil {
				fmt.Printf("Attachment %s of lecture %s downloaded successfully\n", a.Name, l.Title)
			} else {
				fmt.Printf("An error occurred while downloading attachment %s of lecture %s: %v\n", a.Name, l.Title, err)
			}
		}(a)
	}
	
	for i, v := range l.Videos {
		fmt.Printf("Downloading video #%d of lecture %s\n", i+1, l.Title)

		el := l.dlManager.Download(
			v.Name,
			func(client *http.Client) error {
				return v.Download(client, l.Title)
			},
		)

		go func(v *Video) {
			err := el.Wait()
			if err == nil {
				fmt.Printf("Video %s of lecture %s downloaded successfully\n", v.Name, l.Title)
			} else {
				fmt.Printf("An error occurred while downloading video %s of lecture %s: %v\n", v.Name, l.Title, err)
			}
		}(v)
	}
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
