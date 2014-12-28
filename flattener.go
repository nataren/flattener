package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"log"
	"os"
)

func main() {
	type Signature struct {
		XMLName xml.Name `xml:"signature"`
		Value   string   `xml:",innerxml"`
	}

	type IP struct {
		XMLName xml.Name `xml:"ip"`
		Value   string   `xml:",innerxml"`
	}

	type SessionId struct {
		XMLName xml.Name `xml:"session-id"`
		Value   string   `xml:",innerxml"`
	}

	type User struct {
		XMLName   xml.Name `xml:"user"`
		Id        string   `xml:"id,attr"`
		Anonymous string   `xml:"anonymous,attr"`
	}

	type Request struct {
		XMLName   xml.Name  `xml:"request"`
		Id        string    `xml:"id,attr"`
		Seq       string    `xml:"seq,attr"`
		Count     string    `xml:"count,attr"`
		Signature Signature `xml:"signature"`
		IP        IP        `xml:"ip"`
		SessionId SessionId `xml:"session-id"`
		// Parameters []string `xml:"parameters"`
		User User `xml:"user"`
	}

	type PagePath struct {
		XMLName xml.Name `xml:"path"`
		Value   string   `xml:",innerxml"`
	}

	type Page struct {
		XMLName xml.Name `xml:"page"`
		Id      string   `xml:"id,attr"`
		Path    PagePath `xml:"path"`
	}

	type Data struct {
		XMLName   xml.Name `xml:"data"`
		UriHost   string   `xml:"_uri.host"`
		UriScheme string   `xml:"_uri.scheme"`
		UriQuery  string   `xml:"_uri.query"`
	}

	type MindTouchEvent struct {
		XMLName   xml.Name `xml:"event"`
		Id        string   `xml:"id,attr"`
		Datetime  string   `xml:"datetime,attr"`
		Type      string   `xml:"type,attr"`
		Wikiid    string   `xml:"wikiid,attr"`
		Journaled string   `xml:"journaled,attr"`
		Version   string   `xml:"version,attr"`
		Request   Request  `xml:"request"`
		Page      Page     `xml:"page"`
		Data      Data     `xml:"data"`
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
		ev := MindTouchEvent{}
		unmarshalErr := xml.Unmarshal([]byte(event), &ev)
		if unmarshalErr != nil {
			log.Printf("Could not deserialize: '%s'", event)
		}
		fmt.Println(ev)
	}
}
