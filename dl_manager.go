package main

import (
	"net/http"
	"sync"

	"github.com/google/uuid"
)

type dlElement struct {
	dlFunc func() error
	UUID   string
}

type dlManager struct {
	m 	   *sync.Mutex
	wg 	   *sync.WaitGroup
	client *http.Client
	els    map[int]*dlElement
}

func newDLManager(client *http.Client) *dlManager {
	return &dlManager {
		m: new(sync.Mutex),
		wg: new(sync.WaitGroup),
		client: client,
		els: make(map[int]*dlElement),
	}
}

func (dl *dlManager) RegisterDownload(priority int, dlFunc func() error) *dlElement {
	el := &dlElement { UUID: uuid.NewString(), dlFunc: dlFunc }
	for _, ok := dl.els[priority]; ok; priority++ {}
	dl.els[priority] = el

	return el
}
