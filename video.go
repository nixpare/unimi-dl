package main

import (
	"fmt"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Video struct {
	uploadTime  time.Time
	manifestURL string
	size        string
	infos       []string
}

const (
	VIDEO_QUERY = "div.lecturecVideo"
	VIDEO_MANIFEST_QUERY = "video.lecturec > source"
	VIDEO_INFO_QUERY = "div > p"
)

func NewVideo(s *goquery.Selection) Video {
	var v Video
	v.manifestURL = s.Find(VIDEO_MANIFEST_QUERY).AttrOr("src", "")

	v.infos = make([]string, 0)
	s.Find(VIDEO_INFO_QUERY).Each(func(i int, s *goquery.Selection) {
		info := ExtractTextFromHTML(s)
		if info == "" {
			return
		}
		v.infos = append(v.infos, info)
	})

	return v
}

func (v Video) String() string {
	res := fmt.Sprintf("Manifest URL: %s\nInfo:", v.manifestURL)
	if len(v.infos) == 0 {
		return res + " None"
	}

	for _, info := range v.infos {
		res += fmt.Sprintf("\n  -  %s", info)
	}
	return res
}
