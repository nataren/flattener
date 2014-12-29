package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"log"
	"os"
)

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

type Parameter struct {
	XMLName xml.Name `xml:"param"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:",innerxml"`
}

type Request struct {
	XMLName    xml.Name    `xml:"request"`
	Id         string      `xml:"id,attr"`
	Seq        string      `xml:"seq,attr"`
	Count      string      `xml:"count,attr"`
	Signature  Signature   `xml:"signature"`
	IP         IP          `xml:"ip"`
	SessionId  SessionId   `xml:"session-id"`
	Parameters []Parameter `xml:"parameters>param"`
	User       User        `xml:"user"`
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

type File struct {
	XMLName  xml.Name `xml:"file"`
	Id       string   `xml:"id,attr"`
	ResId    string   `xml:"res-id,attr"`
	Filename string   `xml:"filename"`
}

type Data struct {
	XMLName   xml.Name `xml:"data"`
	UriHost   string   `xml:"_uri.host"`
	UriScheme string   `xml:"_uri.scheme"`
	UriQuery  string   `xml:"_uri.query"`
}

type Event struct {
	XMLName   xml.Name `xml:"event"`
	Id        string   `xml:"id,attr"`
	Datetime  string   `xml:"datetime,attr"`
	Type      string   `xml:"type,attr"`
	Cascading string   `xml:"cascading,attr"`
	Wikiid    string   `xml:"wikiid,attr"`
	Journaled string   `xml:"journaled,attr"`
	Version   string   `xml:"version,attr"`
	Request   Request  `xml:"request"`
	IsImage   string   `xml:"isimage"`
	Page      Page     `xml:"page"`
	File      File     `xml:"file"`
	Data      Data     `xml:"data"`
}

var header []string = []string{
	"id",                     // 0
	"datetime",               // 1
	"type",                   // 2
	"cascading",              // 3
	"wikiid",                 // 4
	"journaled",              // 5
	"version",                // 6
	"request.id",             // 7
	"request.seq",            // 8
	"request.count",          // 9
	"request.signature",      // 10
	"request.ip",             // 11
	"request.sessionid",      // 12
	"request.parameters",     // 13
	"request.user.id",        // 14
	"request.user.anonymous", // 15
	"isimage",                // 16
	"page.id",                // 17
	"page.path",              // 18
	"file.id",                // 19
	"file.res-id",            // 20
	"file.filename",          // 21
	"data.urihost",           // 22
	"data.urischeme",         // 23
	"data.uriquery",          // 24
}

func (ev Event) ToStringArray() []string {
	values := make([]string, len(header))

	// Group the optional parameters together as key1:value1;key2:value2...
	var params bytes.Buffer
	firstParam := true
	for _, p := range ev.Request.Parameters {
		if firstParam {
			firstParam = false
		} else {
			params.WriteString(";")
		}
		params.WriteString(p.Name)
		params.WriteString(":")
		params.WriteString(p.Value)
	}

	// Populate the values
	values[0] = ev.Id
	values[1] = ev.Datetime
	values[2] = ev.Type
	values[3] = ev.Cascading
	values[4] = ev.Wikiid
	values[5] = ev.Journaled
	values[6] = ev.Version
	values[7] = ev.Request.Id
	values[8] = ev.Request.Seq
	values[9] = ev.Request.Count
	values[10] = ev.Request.Signature.Value
	values[11] = ev.Request.IP.Value
	values[12] = ev.Request.SessionId.Value
	values[13] = params.String()
	values[14] = ev.Request.User.Id
	values[15] = ev.Request.User.Anonymous
	values[16] = ev.IsImage
	values[17] = ev.Page.Id
	values[18] = ev.Page.Path.Value
	values[19] = ev.File.Id
	values[20] = ev.File.ResId
	values[21] = ev.File.Filename
	values[22] = ev.Data.UriHost
	values[23] = ev.Data.UriScheme
	values[24] = ev.Data.UriQuery
	return values
}

func main() {

	// Open file
	events, err := os.Open("site_1-events-20140206.log")
	if err != nil {
		log.Fatal("Error opening file 'site_1-events-20140206.log")
		return
	}
	defer events.Close()
	r := bufio.NewReader(events)

	// Create new file
	csvFile, err := os.Create("site_1-events-20140206.log.csv")
	if err != nil {
		log.Fatal("Error creating file")
		return
	}
	w := csv.NewWriter(bufio.NewWriter(csvFile))
	w.Write(header)

	// Read all the events
	for {
		event, err := r.ReadString('\x00')
		if err != nil {
			log.Printf("There was an error while reading from the events stream, will exit, '%s'", err)
			break
		}
		if event == "" {
			break
		}
		ev := Event{}
		unmarshalErr := xml.Unmarshal([]byte(event), &ev)
		if unmarshalErr != nil {
			log.Printf("Could not deserialize: '%s'", event)
			continue
		}
		fmt.Printf("Processing event: %s\n", ev.Id)
		w.Write(ev.ToStringArray())
	}
	w.Flush()
}
