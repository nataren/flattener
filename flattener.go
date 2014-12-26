package main

import (
	"encoding/xml"
	"fmt"
	"os"
	"bufio"
	"log"
)

func main() {
	type MindTouchEvent struct {
		XMLName xml.Name `xml:"event"`
		Id  string  `xml:"id,attr"`
	}

	// Open file
	events, err := os.Open("site_1-events-20140206.log")
	if err != nil {
		log.Fatal("Error opening file 'site_1-events-20140206.log")
		return
	}
	defer events.Close()
	r := bufio.NewReader(events)

	// Read all the events
	for {
		event, err := r.ReadString('\x00')
		if err != nil {
			log.Printf("There was an error while reading from the events stream, '%s'", err)
		}
		if event == "" {
			break
		}
		ev := MindTouchEvent { }
        unmarshalErr := xml.Unmarshal([]byte(event), &ev)
        if unmarshalErr != nil {
	        log.Printf("Could not deserialize: '%s'", event)
        }
        fmt.Println(ev)
	}
}