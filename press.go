package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/xml"
	"flag"
	"io"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"
	"sync"
)

func main() {

	// Setup flags
	var (
		dirname    string
		debug      bool
		cpuprofile string
		workers    int
	)
	flag.StringVar(&dirname, "dir", "", "The path to the folder where the log files live")
	flag.BoolVar(&debug, "debug", false, "Enable debug messages, in particular it shows which item has been scheduled for processing")
	flag.StringVar(&cpuprofile, "cpuprofile", "", "Write cpu profile to file")
	flag.IntVar(&workers, "workers", 100, "The limit of concurrent workers at a time")
	flag.Parse()

	// Validate flags
	if dirname == "" {
		log.Println("There's nothing to do, will exit")
		return
	}
	dir, err := os.Open(dirname)
	if err != nil {
		log.Printf("Could not read directory '%s': '%s'", dirname, err)
		return
	}
	dirinfo, err := dir.Readdir(-1)
	if err != nil {
		log.Printf("There was a problem reading the contents of the directory '%s': '%s'", dirname, err)
	}

	// Profiling settings
	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	// Parallelism settings
	cpus := runtime.NumCPU()
	runtime.GOMAXPROCS(cpus)

	// Iteration counters
	start := 0
	increment := workers
	tentativeOffset := increment
	limit := len(dirinfo) - 1
	for {

		// Stop processing once we are done with all the files
		if start >= limit {
			break
		}
		var wg sync.WaitGroup

		// Process chunk
		offset := min(tentativeOffset, limit)
		for _, fi := range dirinfo[start:offset] {
			filename := fi.Name()
			wg.Add(1)
			go func() {
				defer wg.Done()
				if debug {
					log.Printf("Currently processing file '%s'", filename)
				}
				if !strings.HasSuffix(filename, ".log") {
					log.Printf("Will skip file '%s'", filename)
					return
				}

				// Open file
				events, err := os.Open(filename)
				if err != nil {
					log.Printf("Error opening file '%s': %v", filename, err)
					return
				}
				defer events.Close()
				r := bufio.NewReader(events)

				// Create new file
				csvFile, err := os.Create(filename + ".csv")
				if err != nil {
					log.Printf("Error creating csv file for '%s'", filename)
					return
				}
				w := csv.NewWriter(bufio.NewWriter(csvFile))
				w.Write(header)

				// Read all the events
				for {
					event, err := r.ReadString('\x00')
					if err != nil {
						if err == io.EOF {
							if debug {
								log.Printf("Finished processing '%s'", filename)
							}
							break
						}
						log.Printf("There was an error while reading from the events stream, will exit, '%s'", err)
						break
					}
					if event == "" {
						break
					}
					ev := Event{}
					unmarshalErr := xml.Unmarshal([]byte(event), &ev)
					if unmarshalErr != nil {
						log.Printf("Could not deserialize: '%s', '%s', '%s'", filename, event, unmarshalErr)
						continue
					}
					if ev.Any != "" {
						log.Printf("'%s': The event '%s' was found not having all its members deserialized. Any = '%s'", event, filename, ev.Any)
					}
					w.Write(ev.ToStringArray())
				}
				w.Flush()
				return
			}()
		}
		wg.Wait()
		start += increment
		tentativeOffset += increment
	}
	log.Printf("used %d CPUs", cpus)
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

	// Tags added
	var tagsAdded bytes.Buffer
	firstTag := true
	for _, tagAdded := range ev.TagsAdded {
		if firstTag {
			firstTag = false
		} else {
			tagsAdded.WriteString(";")
		}
		tagsAdded.WriteString(tagAdded.Name)
		tagsAdded.WriteString("^")
		tagsAdded.WriteString(tagAdded.Type)
	}

	// Tags removed
	var tagsRemoved bytes.Buffer
	firstRemovedTag := true
	for _, tagRemoved := range ev.TagsRemoved {
		if firstRemovedTag {
			firstRemovedTag = false
		} else {
			tagsRemoved.WriteString(";")
		}
		tagsRemoved.WriteString(tagRemoved.Name)
		tagsRemoved.WriteString("^")
		tagsRemoved.WriteString(tagRemoved.Type)
	}

	// Populate the values
	values[0] = ev.ID
	values[1] = ev.Datetime
	values[2] = ev.Type
	values[3] = ev.Cascading
	values[4] = ev.Wikiid
	values[5] = ev.Journaled
	values[6] = ev.Version
	values[7] = ev.Request.ID
	values[8] = ev.Request.Seq
	values[9] = ev.Request.Count
	values[10] = ev.Request.Signature.Value
	values[11] = ev.Request.IP.Value
	values[12] = ev.Request.SessionID.Value
	values[13] = params.String()
	values[14] = ev.Request.User.ID
	values[15] = ev.Request.User.Anonymous
	values[16] = ev.IsImage
	values[17] = ev.Page.ID
	values[18] = ev.Page.Path.Value
	values[19] = ev.File.ID
	values[20] = ev.File.ResID
	values[21] = ev.File.Filename
	values[22] = ev.Data.URIHost
	values[23] = ev.Data.URIScheme
	values[24] = ev.Data.URIQuery
	values[25] = ev.Diff.Added
	values[26] = ev.Diff.Removed
	values[27] = ev.Diff.Attributes
	values[28] = ev.Diff.Structural
	values[29] = ev.CreateReason
	values[30] = ev.User.ID
	values[31] = ev.User.Name
	values[32] = ev.CreateReasonDetail
	values[33] = ev.DescendantPage.ID
	values[34] = ev.DescendantPage.Path.Value
	values[35] = ev.RootCopyPage.ID
	values[36] = ev.RootCopyPage.Path.Value
	values[37] = ev.RootDeletePage.ID
	values[38] = ev.RootDeletePage.Path.Value
	values[39] = ev.RootPage.ID
	values[40] = ev.RootPage.Path.Value
	values[41] = ev.SourcePage.ID
	values[42] = ev.SourcePage.Path.Value
	values[43] = ev.From
	values[44] = ev.To
	values[45] = ev.Revision
	values[46] = ev.RevisionPrevious
	values[47] = ev.RevisionReverted
	values[48] = ev.Comment.ID
	values[49] = ev.Comment.Content.Type
	values[50] = ev.Comment.Content.Value
	values[51] = tagsAdded.String()
	values[52] = tagsRemoved.String()
	values[53] = ev.Property.ID
	values[54] = ev.Property.Name
	values[55] = ev.RestrictionID
	values[56] = ev.PreviousRestrictionID
	values[57] = ev.Score
	values[58] = ev.Grant.ID
	values[59] = ev.Grant.Type
	values[60] = ev.Grant.Role.ID
	values[61] = ev.Grant.User.ID
	values[62] = ev.Any
	values[63] = ev.User.Username
	values[64] = ev.AuthMethodPassword
	values[65] = ev.AuthMethodApikey
	values[66] = ev.Origin
	values[67] = ev.ReasonType
	values[68] = ev.DisplayNamePrevious
	values[69] = ev.DisplayNameCurrent
	values[70] = ev.ContentTypePrevious
	values[71] = ev.ContentTypeCurrent
	values[72] = ev.ChangeComment
	values[73] = ev.TitleSegmentPrevious
	values[74] = ev.TitleSegmentCurrent
	values[75] = ev.Data.Query
	values[76] = ev.Data.Constraint
	values[77] = ev.Workflow.Name
	values[78] = ev.Workflow.URINext
	values[79] = ev.Workflow.Data.UserID
	values[80] = ev.Workflow.Data.Username
	values[81] = ev.Workflow.Data.CustomerActivityID
	values[82] = ev.Workflow.Data.RequestID
	values[83] = ev.Workflow.Data.Email
	values[84] = ev.Workflow.Data.Search
	values[85] = ev.Workflow.Data.Path
	return values
}

var header = []string{
	"id",                                // 0
	"datetime",                          // 1
	"type",                              // 2
	"cascading",                         // 3
	"wikiid",                            // 4
	"journaled",                         // 5
	"version",                           // 6
	"request.id",                        // 7
	"request.seq",                       // 8
	"request.count",                     // 9
	"request.signature",                 // 10
	"request.ip",                        // 11
	"request.sessionid",                 // 12
	"request.parameters",                // 13
	"request.user.id",                   // 14
	"request.user.anonymous",            // 15
	"isimage",                           // 16
	"page.id",                           // 17
	"page.path",                         // 18
	"file.id",                           // 19
	"file.resid",                        // 20
	"file.filename",                     // 21
	"data._urihost",                     // 22
	"data._urischeme",                   // 23
	"data._uriquery",                    // 24
	"diff.added",                        // 25
	"diff.removed",                      // 26
	"diff.attributes",                   // 27
	"diff.structural",                   // 28
	"createreason",                      // 29
	"user.id",                           // 30
	"user.name",                         // 31
	"createreasondetail",                // 32
	"descendant.page.id",                // 33
	"descendant.page.path",              // 34
	"root.copy.page.id",                 // 35
	"root.copy.page.path",               // 36
	"root.delete.page.id",               // 37
	"root.delete.page.path",             // 38
	"root.page.id",                      // 39
	"root.page.path",                    // 40
	"source.page.id",                    // 41
	"source.page.path",                  // 42
	"from",                              // 43
	"to",                                // 44
	"revision",                          // 45
	"revision.previous",                 // 46
	"revision.reverted",                 // 47
	"comment.id",                        // 48
	"comment.content.type",              // 49
	"comment.content",                   // 50
	"tags.added",                        // 51
	"tags.removed",                      // 52
	"property.id",                       // 53
	"property.name",                     // 54
	"restriction.id",                    // 55
	"previous.restriction.id",           // 56
	"score",                             // 57
	"grant.id",                          // 58
	"grant.type",                        // 59
	"grant.role.id",                     // 60
	"grant.user.id",                     // 61
	"any",                               // 62
	"user.username",                     // 63
	"authmethodpassword",                // 64
	"authmethodapikey",                  // 65
	"origin",                            // 66
	"reasontype",                        // 67
	"displayname.previous",              // 68
	"displayname.current",               // 69
	"contenttype.previous",              // 70
	"contenttype.current",               // 71
	"change-comment",                    // 72
	"titlesegment.previous",             // 73
	"titlesegment.current",              // 74
	"data.query",                        // 75
	"data.constraint",                   // 76
	"workflow.name",                     // 77
	"workflow.urinext",                  // 78
	"workflow.data._userid",             // 79
	"workflow.data._username",           // 80
	"workflow.data._customeractivityid", // 81
	"workflow.data._requestid",          // 82
	"workflow.data._email",              // 83
	"workflow.data._search",             // 84
	"workflow.data._path",               // 85
}

type Event struct {
	XMLName               xml.Name       `xml:"event"`
	ID                    string         `xml:"id,attr"`
	Datetime              string         `xml:"datetime,attr"`
	Type                  string         `xml:"type,attr"`
	Cascading             string         `xml:"cascading,attr"`
	Wikiid                string         `xml:"wikiid,attr"`
	Journaled             string         `xml:"journaled,attr"`
	Version               string         `xml:"version,attr"`
	Request               Request        `xml:"request"`
	IsImage               string         `xml:"isimage"`
	Page                  Page           `xml:"page"`
	File                  File           `xml:"file"`
	Data                  Data           `xml:"data"`
	Diff                  Diff           `xml:"diff"`
	CreateReason          string         `xml:"create-reason"`
	User                  User           `xml:"user"`
	CreateReasonDetail    string         `xml:"create-reason-detail"`
	DescendantPage        DescendantPage `xml:"descendant.page"`
	RootCopyPage          RootCopyPage   `xml:"root.copy.page"`
	RootDeletePage        RootDeletePage `xml:"root.delete.page"`
	RootPage              RootPage       `xml:"root.page"`
	SourcePage            SourcePage     `xml:"source.page"`
	From                  string         `xml:"from"`
	To                    string         `xml:"to"`
	Revision              string         `xml:"revision"`
	RevisionPrevious      string         `xml:"revision.previous"`
	RevisionReverted      string         `xml:"revision.reverted"`
	Comment               Comment        `xml:"comment"`
	TagsAdded             []Tag          `xml:"tags-added>tag"`
	TagsRemoved           []Tag          `xml:"tags-removed>tag"`
	Property              Property       `xml:"property"`
	RestrictionID         string         `xml:"restriction-id"`
	PreviousRestrictionID string         `xml:"previous.restriction-id"`
	Score                 string         `xml:"score"`
	Grant                 Grant          `xml:"grant"`
	Any                   string         `xml:",any"`
	AuthMethodPassword    string         `xml:"authmethod-password"`
	AuthMethodApikey      string         `xml:"authmethod-apikey"`
	Origin                string         `xml:"origin"`
	ReasonType            string         `xml:"reason-type"`
	DisplayNamePrevious   string         `xml:"displayname.previous"`
	DisplayNameCurrent    string         `xml:"displayname.current"`
	ContentTypePrevious   string         `xml:"contenttype.previous"`
	ContentTypeCurrent    string         `xml:"contenttype.current"`
	ChangeComment         string         `xml:"change-comment"`
	TitleSegmentPrevious  string         `xml:"titlesegment.previous"`
	TitleSegmentCurrent   string         `xml:"titlesegment.current"`
	Workflow              Workflow       `xml:"workflow"`
	UpdateReason          string         `xml:"update-reason"`
}

type Signature struct {
	XMLName xml.Name `xml:"signature"`
	Value   string   `xml:",innerxml"`
}

type IP struct {
	XMLName xml.Name `xml:"ip"`
	Value   string   `xml:",innerxml"`
}

type SessionID struct {
	XMLName xml.Name `xml:"session-id"`
	Value   string   `xml:",innerxml"`
}

type User struct {
	XMLName   xml.Name `xml:"user"`
	ID        string   `xml:"id,attr"`
	Anonymous string   `xml:"anonymous,attr"`
	Name      string   `xml:"name"`
	Username  string   `xml:"username,attr"`
}

type Parameter struct {
	XMLName xml.Name `xml:"param"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:",innerxml"`
}

type Request struct {
	XMLName    xml.Name    `xml:"request"`
	ID         string      `xml:"id,attr"`
	Seq        string      `xml:"seq,attr"`
	Count      string      `xml:"count,attr"`
	Signature  Signature   `xml:"signature"`
	IP         IP          `xml:"ip"`
	SessionID  SessionID   `xml:"session-id"`
	Parameters []Parameter `xml:"parameters>param"`
	User       User        `xml:"user"`
}

type PagePath struct {
	XMLName xml.Name `xml:"path"`
	Value   string   `xml:",innerxml"`
}

type Page struct {
	XMLName xml.Name `xml:"page"`
	ID      string   `xml:"id,attr"`
	Path    PagePath `xml:"path"`
}

type RootPage struct {
	XMLName xml.Name `xml:"root.page"`
	ID      string   `xml:"id,attr"`
	Path    PagePath `xml:"path"`
}

type SourcePage struct {
	XMLName xml.Name `xml:"source.page"`
	ID      string   `xml:"id,attr"`
	Path    PagePath `xml:"path"`
}

type DescendantPage struct {
	XMLName xml.Name `xml:"descendant.page"`
	ID      string   `xml:"id,attr"`
	Path    PagePath `xml:"path"`
}

type RootCopyPage struct {
	XMLName xml.Name `xml:"root.copy.page"`
	ID      string   `xml:"id,attr"`
	Path    PagePath `xml:"path"`
}

type RootDeletePage struct {
	XMLName xml.Name `xml:"root.delete.page"`
	ID      string   `xml:"id,attr"`
	Path    PagePath `xml:"path"`
}

type File struct {
	XMLName  xml.Name `xml:"file"`
	ID       string   `xml:"id,attr"`
	ResID    string   `xml:"res-id,attr"`
	Filename string   `xml:"filename"`
}

type Data struct {
	XMLName    xml.Name `xml:"data"`
	URIHost    string   `xml:"_uri.host"`
	URIScheme  string   `xml:"_uri.scheme"`
	URIQuery   string   `xml:"_uri.query"`
	Query      string   `xml:"query"`
	Constraint string   `xml:"constraint"`
}

type Diff struct {
	XMLName    xml.Name `xml:"diff"`
	Added      string   `xml:"added"`
	Removed    string   `xml:"removed"`
	Attributes string   `xml:"attributes"`
	Structural string   `xml:"structural"`
}

type CommentContent struct {
	XMLName xml.Name `xml:"content"`
	Type    string   `xml:"type,attr"`
	Value   string   `xml:",innerxml"`
}

type Comment struct {
	XMLName xml.Name       `xml:"comment"`
	ID      string         `xml:"id,attr"`
	Content CommentContent `xml:"content"`
}

type Tag struct {
	XMLName xml.Name `xml:"tag"`
	Name    string   `xml:"name"`
	Type    string   `xml:"type"`
}

type Property struct {
	XMLName xml.Name `xml:"property"`
	ID      string   `xml:"id"`
	Name    string   `xml:"name"`
}

type Role struct {
	XMLName xml.Name `xml:"role"`
	ID      string   `xml:"id,attr"`
}

type Grant struct {
	XMLName xml.Name `xml:"grant"`
	ID      string   `xml:"id"`
	Type    string   `xml:"type"`
	Role    Role     `xml:"role"`
	User    User     `xml:"user"`
}

type WorkflowData struct {
	XMLName            xml.Name `xml:"data"`
	UserID             string   `xml:"_userid"`
	Username           string   `xml:"_username"`
	CustomerActivityID string   `xml:"_customeractivityid"`
	RequestID          string   `xml:"_requestid"`
	Email              string   `xml:"_email"`
	Search             string   `xml:"_search"`
	Path               string   `xml:"_path"`

	// TODO: There can be custom elements to deserialize. We should use the <param name="{name}">{value}</param> for the representation instead
}

type Workflow struct {
	XMLName xml.Name     `xml:"workflow"`
	Name    string       `xml:"name,attr"`
	URINext string       `xml:"uri.next"`
	Data    WorkflowData `xml:"data"`
}

func min(x, y int) int {
	if x <= y {
		return x
	}
	return y
}
