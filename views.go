package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"strings"
)

type handler struct {
	Function func(http.ResponseWriter, *http.Request)
	Methods  []string
}

var BlocklistsHandler = &handler{
	Function: Blocklists,
	Methods:  []string{"POST"},
}

func Blocklists(w http.ResponseWriter, r *http.Request) {
	_, ids404 := search(r, "block")
	if len(ids404) > 0 {
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

var BlocksHandler = &handler {
    Function: Blocks,
    Methods: []string{"POST"},
}

func Blocks(w http.ResponseWriter, r *http.Request) {
    _, ids404 := search(r, "block")
	if len(ids404) > 0 {
		w.WriteHeader(http.StatusNotFound)
		body, err := json.Marshal(ids404)
		if err != nil {
			fmt.Println(err.Error())
		}
		w.Write(body)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func search(r *http.Request, group string) ([]string, []string) {

	var (
		wg      sync.WaitGroup
		buckets []string
	)

		switch group {
	case "blocklist":
		buckets = blocklistBuckets
	case "blocks":
		buckets = blockBuckets
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err.Error())
	}

	ids := make([]string, 0, 20)
	err = json.Unmarshal(body, &ids)
	if err != nil {
		fmt.Println(err.Error())
	}

	expedition := NewExpedition(len(buckets))

	request := make(chan string, len(ids) * 3)
	response := make(chan string, len(ids) * 3)

	for i := 0; i < num_expeditions; i++ {
		wg.Add(1)
		go expedition.Explore(group, request, response, &wg)
	}

	receivedIds := make(map[string]bool)
	for _, id := range ids {
		request <- id
		receivedIds[id] = false
	}

	// closing this channel will signal to the expeditions that
	// they should exit when done processing their loads.
	close(request)

	// wait for all the expeditions to finish
	wg.Wait()

	// close so we don't hang on the expedition response
	close(response)

	for id := range response {
		parts := strings.Split(id, ":")
		id = parts[0]
		if _, ok := receivedIds[id]; ok {
			receivedIds[id] = true
		}
	}

	ids200 := make([]string, 0, len(receivedIds))
	ids404 := make([]string, 0, len(receivedIds))

	for k, v := range receivedIds {
		if v {
			ids200 = append(ids200, k)
		} else {
			ids404 = append(ids404, k)
		}
	}

	return ids200, ids404
}
