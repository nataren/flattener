package main

import (
	"encoding/xml"
	"fmt"
	"os"
	"bufio"
	"log"
)

func main() {
	type Signature struct {
		SignatureString string `xml:",innerxml"`
	}

	type IP struct {
		IPString string `xml:",innerxml"`
	}

	type SessionId struct {
		SessionIdString string `xml:",innerxml"`
	}

	type Request struct {
		Id string `xml:"id,attr"`
		Seq string `xml:"seq,attr"`
		Count string `xml:"count,attr"`
		Signature Signature `xml:"signature"`
		IP IP `xml:"ip"`
		SessionId SessionId `xml:"session-id"`
		// Parameters []string `xml:"request>parameters"`
	}

	type MindTouchEvent struct {
		XMLName xml.Name `xml:"event"`
		Id  string  `xml:"id,attr"`
		Datetime string `xml:"datetime,attr"`
		Type string `xml:"type,attr"`
		Wikiid string `xml:"wikiid,attr"`
		Journaled string `xml:"journaled,attr"`
		Version string `xml:"version,attr"`
		Request Request `xml:"request"`
		// RequestUser
		// PageId
		// PagePath string `xml:"page>path"`
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