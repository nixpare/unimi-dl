package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type dlFunc func(client *http.Client) error

type DLElement struct {
	Name        string
	dlF         dlFunc
	DLM         *DLManager
	Downloading bool
	startChan   chan struct{}
	errChan     chan error
}

func (el *DLElement) download() error {
	el.DLM.startDownload(el)
	
	err := el.dlF(el.DLM.Client)
	el.DLM.downloadFinished(el)

	el.DLM.startNextDownload()

	el.errChan <- err
	return err
}

func (el *DLElement) Wait() error {
	return <- el.errChan
} 

type DLManager struct {
	m 	         *sync.Mutex
	wg 	         *sync.WaitGroup
	Client       *http.Client
	els	         map[string]*DLElement
	count        int
	MaxDownloads int
}

func NewDLManager(client *http.Client) *DLManager {
	return &DLManager {
		m: new(sync.Mutex),
		wg: new(sync.WaitGroup),
		Client: client,
		els: make(map[string]*DLElement),
		MaxDownloads: 4,
	}
}

func (dlm *DLManager) startDownload(el *DLElement) {
	dlm.m.Lock()
	if dlm.count < dlm.MaxDownloads {
		el.startChan <- struct{}{}
	}
	dlm.m.Unlock()

	<- el.startChan

	dlm.m.Lock()
	defer dlm.m.Unlock()

	dlm.count ++
	el.Downloading = true
}

func (dlm *DLManager) downloadFinished(el *DLElement) {
	dlm.m.Lock()
	defer dlm.m.Unlock()

	dlm.count --
	delete(dlm.els, el.Name)
	el.Downloading = false
	dlm.wg.Done()
}

func (dlm *DLManager) startNextDownload() {
	dlm.m.Lock()
	defer dlm.m.Unlock()
	
	for _, el := range dlm.els {
		if !el.Downloading {
			el.startChan <- struct{}{}
			break
		}
	}

	fmt.Printf("\nDownloading (%d):\n", dlm.count)
	for _, el := range dlm.els {
		if el.Downloading {
			fmt.Println(el.Name)
		}
	}

	fmt.Printf("\nDownload pending:\n")
	for _, el := range dlm.els {
		if !el.Downloading {
			fmt.Println(el.Name)
		}
	}
	
	fmt.Println()
}

func (dlm *DLManager) Download(name string, f dlFunc) *DLElement {
	dlm.m.Lock()
	defer dlm.m.Unlock()

	dlm.wg.Add(1)
	el := &DLElement {
		Name: name,
		dlF: f,
		DLM: dlm,
		startChan: make(chan struct{}, 1),
		errChan: make(chan error, 1),
	}
	dlm.els[name] = el

	go el.download()
	return el
}

func (dlm *DLManager) WaitForAllDownloads() {
	time.Sleep(time.Second)
	dlm.wg.Wait()
}
