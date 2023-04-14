package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

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
	fmt.Printf("Downloading video <%s> with url <%s>\n", v.Name, v.manifestURL)
	dirPath, _ := os.Getwd()

	prefix = strings.ReplaceAll(prefix, "/", "-")
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	videoPath := dirPath + "/" + prefix + strings.ReplaceAll(v.Name, "/", "-") + ".ts"
	
	videoDownload, err := v.newVideoDownload(videoPath)
	if err != nil {
		return err
	}

	defer fmt.Printf("Video <%s> downloaded\n", v.Name)
	return videoDownload.download()
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
	flushM   *sync.Mutex
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
		flushM: new(sync.Mutex),
	}, nil
}

func (vd *videoDownload) downloadSegment(index int, segURL string) error {
	data, err := getRequest(vd.client, segURL)
	if err != nil {
		return err
	}

	vs := vd.segments[index]
	vs.data = data
	vs.ready = true

	vd.doneChan <- index
	return nil
}

func (vd *videoDownload) flush(writeC chan <- struct{ start int; end int }, writeWG *sync.WaitGroup) {
	vd.flushM.Lock()
	defer vd.flushM.Unlock()

	sliceStart := -1
	for i := 0; i < len(vd.segments); i++ {
		videoSeg := vd.segments[i]
		if videoSeg == nil || !videoSeg.ready {
			sliceStart = -1
			break
		}

		if videoSeg.ready && !videoSeg.written {
			sliceStart = i
			break
		}
	}

	if sliceStart == -1 {
		return
	}

	sliceEnd := -1
	for i := sliceStart + 1; i < len(vd.segments); i++ {
		videoSeg := vd.segments[i]
		if videoSeg == nil || !videoSeg.ready || videoSeg.written {
			break
		}
		sliceEnd = i
	}

	if sliceEnd == -1 {
		return
	}

	writeWG.Add(1)
	writeC <- struct{ start int; end int }{ start: sliceStart, end: sliceEnd }
}

func (vd *videoDownload) download() error {
	manifestContent, err := getRequest(vd.client, vd.v.manifestURL)
	if err != nil {
		return err
	}

	videoBaseURL := vd.v.manifestURL[:strings.LastIndex(vd.v.manifestURL, "/")]

	segManifestSplit := strings.Split(strings.TrimSpace(string(manifestContent)), "\n")
	segManifestURL := videoBaseURL + "/" + segManifestSplit[len(segManifestSplit)-1]

	segManifestContent, err := getRequest(vd.client, segManifestURL)
	if err != nil {
		return err
	}
	
	segsURL := make([]string, 0)
	for _, line := range strings.Split(string(segManifestContent), "\n") {
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		segsURL = append(segsURL, videoBaseURL + "/" + line)
	}
	numSegs := len(segsURL)

	segWG := new(sync.WaitGroup)
	segWG.Add(numSegs)

	segRoutineF := func(index int, segURL string) {
		defer segWG.Done()

		err := vd.downloadSegment(index, segURL)
		if err != nil {
			panic(err)
		}
	}

	writeC := make(chan struct { start int; end int })
	writeWG := new(sync.WaitGroup)
	go func() {
		for a := range writeC {
			fmt.Printf("Flushing from %d to %d\n", a.start+1, a.end+1)

			for i := a.start; i <= a.end; i++ {
				videoSeg := vd.segments[i]

				vd.file.Write(videoSeg.data)
				videoSeg.data = nil
				videoSeg.written = true
			}
			writeWG.Done()
		}
	}()

	for i, segURL := range segsURL {
		vd.segments[i] = &videoSegment{}

		if i < CONCURRENT_DOWNLOADS {
			go segRoutineF(i, segURL)
			continue
		}

		<- vd.doneChan

		go segRoutineF(i, segURL)
		vd.flush(writeC, writeWG)
	}

	segWG.Wait()
	vd.flush(writeC, writeWG)

	writeWG.Wait()
	close(writeC)
	vd.file.Close()

	return nil
}
