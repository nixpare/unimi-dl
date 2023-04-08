package main

import (
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	VIDEO_QUERY = "div.lecturecVideo"
	VIDEO_MANIFEST_QUERY = "video.lecturec > source"
	VIDEO_INFO_QUERY = "div > p"
)

type Video struct {
	Name 		string
	manifestURL string
	Infos       []string
}

func newVideo(name string, s *goquery.Selection) *Video {
	v := &Video {
		Name: name,
		manifestURL: s.Find(VIDEO_MANIFEST_QUERY).AttrOr("src", ""),
		Infos: make([]string, 0),
	}

	s.Find(VIDEO_INFO_QUERY).Each(func(i int, s *goquery.Selection) {
		info := ExtractTextFromHTML(s)
		if info == "" {
			return
		}
		v.Infos = append(v.Infos, info)
	})

	return v
}

func (v Video) Download(client *http.Client, prefix string) error {
	failErr := fmt.Errorf("an error has occurred while downloading the video %s", v.Name)

	resp, err := client.Get(v.manifestURL)
	if err != nil || resp.StatusCode >= 400 {
		log.Printf("Error getting manifest for video %s (Code %d): %v\n", v.Name, resp.StatusCode, err)
		return failErr
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading manifest response for video %s: %v\n", v.Name, err)
		return failErr
	}

	videoBaseURL := v.manifestURL[:strings.LastIndex(v.manifestURL, "/")]
	dirPath, _ := os.Getwd()

	prefix = strings.ReplaceAll(prefix, "/", "-")
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	videoPath := dirPath + "/" + prefix + strings.ReplaceAll(v.Name, "/", "-") + ".ts"
	videoFile, err := os.OpenFile(videoPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0777)
	if err != nil {
		log.Printf("Error creating file for video %s: %v\n", v.Name, err)
		return failErr
	}
	defer videoFile.Close()

	segManifestSplit := strings.Split(strings.TrimSpace(string(respBody)), "\n")
	segManifestURL := videoBaseURL + "/" + segManifestSplit[len(segManifestSplit)-1]

	resp, err = http.Get(segManifestURL)
	if err != nil {
		log.Printf("Error getting segments manifest for video %s: %v\n", v.Name, err)
		return failErr
	}

	respBody, err = io.ReadAll(resp.Body)
	if err != nil{
		log.Printf("Error reading segments manifest response for video %s: %v\n", v.Name, err)
		return failErr
	}

	segsURL := make([]string, 0)
	for _, line := range strings.Split(string(respBody), "\n") {
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		segsURL = append(segsURL, videoBaseURL + "/" + line)
	}

	for i, segURL := range segsURL {
		resp, err = http.Get(segURL)
		if err != nil || resp.StatusCode >= 400 {
			log.Printf("Error downloading part %d of video %s (Code: %d): %v\n", i+1, v.Name, resp.StatusCode, err)
			return failErr
		}

		respBody, err = io.ReadAll(resp.Body)
		if err != nil{
			log.Printf("Error reading response of video %s part %d: %v\n", v.Name, i, err)
			return failErr
		}

		videoFile.Write(respBody)
	}

	return nil
}

func (v Video) String() string {
	res := fmt.Sprintf("Manifest URL: %s\nInfo:", v.manifestURL)
	if len(v.Infos) == 0 {
		return res + " None"
	}

	for _, info := range v.Infos {
		res += fmt.Sprintf("\n  - %s", info)
	}
	return res
}

const (
	CONCURRENT_DOWNLOADS = 8
	FLUSH_AFTER = 20
)

type videoSegment struct {
	data  	[]byte
	ready 	bool
	written bool
}

type videoDownload struct {
	v        *Video
	client   *http.Client
	segments map[int]*videoSegment
	file     *os.File
	doneChan chan int
}

func (v *Video) newVideoDownload(filePath string) (*videoDownload, error) {
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0777)
	if err != nil {
		return nil, err
	}

	return &videoDownload {
		v: v,
		segments: make(map[int]*videoSegment),
		file: f,
		doneChan: make(chan int, CONCURRENT_DOWNLOADS),
	}, nil
}

func (vd *videoDownload) downloadSegment(index int, segURL string) {
	response, err := vd.client.Get(segURL)
	if err != nil {
		panic(fmt.Errorf("error downloading video <%s> on segment %d: %w", vd.v.Name, index, err))
	}

	if response.StatusCode >= 400 {
		panic(fmt.Errorf(
			"error downloading video <%s> on segment %d: %w", vd.v.Name, index,
			fmt.Errorf("invalid http response %d", response.StatusCode),
		))
	}

	vs := vd.segments[index]
	vs.data, _ = io.ReadAll(response.Body)
	vs.ready = true

	vd.doneChan <- index
}

func (vd *videoDownload) flush() {
	sliceStart := -1
	for i := 0; i < len(vd.segments); i++ {
		if vd.segments[i].ready && !vd.segments[i].written {
			sliceStart = i
			break
		}
	}

	if sliceStart == -1 {
		return
	}

	sliceEnd := sliceStart
	for i := 0; i < len(vd.segments); i++ {
		if !vd.segments[i].ready {
			break
		}
		sliceEnd = i
	}

	flushAfter := int(math.Max(FLUSH_AFTER, CONCURRENT_DOWNLOADS))
	if sliceEnd - sliceStart < flushAfter - 1 {
		return
	}

	for i := sliceStart; i <= sliceEnd; i++ {
		vs := vd.segments[i]

		vd.file.Write(vs.data)
		vs.data = nil
		vs.written = true
	}
}

func (vd *videoDownload) download() error {
	fmt.Printf("Downloading video <%s> with url <%s>\n", vd.v.Name, vd.v.manifestURL)

	return nil
}
