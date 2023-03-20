package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
)

type video struct {
	name string
	link []string
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	dir, _ := os.ReadDir(".")
	var found bool

	pathDir := path.Dir(strings.ReplaceAll(os.Args[0], "\\", "/"))

	htmlFiles := make(map[string]string)
	dirVideo := make(map[string]string)

	for _, x := range dir {
		if strings.Contains(x.Name(), ".html") {
			found = true
			htmlFiles[x.Name()] = pathDir + "/" + x.Name()
			videoDir := strings.Split(pathDir + "/" + x.Name(), ".")[0]
			os.MkdirAll(videoDir, 0777)
			dirVideo[x.Name()] = videoDir
		}
	}

	if !found {
		fmt.Println("Nessun file html presente nella cartella, metterne uno e riavviare il programma\nPremere invio per chiudere la schermata ...")
		scanner.Scan()
		return
	}

	vg := new(sync.WaitGroup)
	stringErr := new(string)

	for x := range htmlFiles {
		singleFile(x, dirVideo[x], vg, stringErr, scanner)
	}

	vg.Wait()

	fmt.Println("Il programma ha terminato. Se presenti, gli errori verranno mostrati qui sotto.\nPremere invio per chiudere la schermata ...")
	fmt.Printf("\nERRORI:\n")
	if len(*stringErr) == 0 {
		fmt.Printf("\tNessun Errore\n")
	} else {
		fmt.Printf("\n%s\n", *stringErr)
	}
	
	scanner.Scan()
}

func singleFile(fileName, dirPath string, vg *sync.WaitGroup, stringErr *string, scanner *bufio.Scanner) {
	file, _ := os.Open(fileName)
	fileScanner := bufio.NewScanner(file)
	for fileScanner.Scan() {
		if strings.Contains(fileScanner.Text(), "<title>") {
			fileScanner.Scan()
			break
		}
	}

	fmt.Printf("Pagina Web trovata: %s\n", fileName)

	videos := make([]video, 0)

	file:for fileScanner.Scan() {
		if strings.Contains(fileScanner.Text(), "title=\"Lecture Captures\"") {
			parsedLine := strings.Split(fileScanner.Text(), "<span>")
			var videoTitle string
			if index := strings.Index(parsedLine[len(parsedLine)-1], "</span>"); index != -1 {
				videoTitle = parsedLine[len(parsedLine)-1][:index]
			}
			videoLinks := make([]string, 0)
			for fileScanner.Scan() {
				if strings.Contains(fileScanner.Text(), "<video") {
					fileScanner.Scan()
					parsedLine = strings.Split(fileScanner.Text(), "\"")
					for _, x := range parsedLine {
						if strings.Contains(x, "https") {
							videoLinks = append(videoLinks, x)
							break
						}
					}
				}
				if strings.Contains(fileScanner.Text(), "</div></td>") {
					videos = append(videos, video{
						videoTitle, videoLinks,
					})
					continue file
				}
			}
		}
	}
	
	fmt.Println("Trovati questi video:")

	for _, x := range videos {
		fmt.Println("Titolo: ", x.name)
		fmt.Println()
		for _, s := range x.link {
			fmt.Println("\tLink: ", s)
		}
		fmt.Print("\nVuoi Scaricarli? [si/no | s/n] -> ")
		scanner.Scan()
		fmt.Print("\n\n")
		if strings.ToLower(scanner.Text()) == "si" || strings.ToLower(scanner.Text()) == "s" {
			if len(x.link) > 1 {
				for i := range x.link {
					vg.Add(1)
					go singleVideo(x.link[i], x.name + fmt.Sprintf("_%d", i+1), dirPath, vg, stringErr)
				}
			}
			if len(x.link) == 1 {
				vg.Add(1)
				go singleVideo(x.link[0], x.name, dirPath, vg, stringErr)
			}			
		}
	}
}

func singleVideo(link, name, dirPath string, vg *sync.WaitGroup, stringErr *string) {
	defer vg.Done()
	failCount := 0

	resp, err := http.Get(link)
	for err != nil {
		failCount ++
		if failCount > 10 {
			*stringErr += fmt.Sprintf("Errore dal video %s con link %s: %s\n", name, link, err.Error())
			return
		}
		resp, err = http.Get(link)
	}

	failCount = 0

	respBody, err := io.ReadAll(resp.Body)
	if err != nil{
		*stringErr += fmt.Sprintf("Errore video \"%s\" con link \"%s\": %v\n", name, link, err.Error())
		return
	}

	parsedVideo := strings.SplitAfter(string(respBody), "chunklist_")
	if len(parsedVideo) < 2 {
		*stringErr += fmt.Sprintf("Errore video \"%s\" con link \"%s\": %v\n", name, link, "impossibile trovare il video")
		return
	}

	parsedUrl := strings.Split(link, "manifest.m3u8")
	if len(parsedUrl) != 2 {
		*stringErr += fmt.Sprintf("Errore video \"%s\" con link \"%s\": %v\n", name, link, "impossibile fare il parsing dell'url")
		return
	}

	videoUrl := parsedUrl[0] + parsedVideo[1][:strings.Index(parsedVideo[1], ".")]
	var index int
	videoFile, err := os.OpenFile(dirPath + "/" + strings.Replace(name, "/", "-", -1)+".ts", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		*stringErr += fmt.Sprintf("Errore video \"%s\" con link \"%s\": %v\n", name, link, err.Error())
		return
	}
	defer videoFile.Close()

	for {
		resp, err = http.Get(videoUrl + fmt.Sprintf("_%d.ts", index))
		for err != nil {
			failCount ++
			if failCount > 10 {
				*stringErr += fmt.Sprintf("Errore video \"%s\" con link \"%s\": %v\n", name, link, "impossibile scaricare il video")
				return
			}
			resp, err = http.Get(videoUrl + fmt.Sprintf("_%d.ts", index))
		}
		respBody, err = io.ReadAll(resp.Body)
		if err != nil{
			*stringErr += fmt.Sprintf("Errore video \"%s\" con link \"%s\": %v\n", name, link, err.Error())
			return
		}
		if len(respBody) == 0 {
			break
		}
		videoFile.Write(respBody)

		index ++
	}
}