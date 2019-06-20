package main

import (
	"bytes"
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"log"
	"os"

	_ "github.com/denisenkom/go-mssqldb"
)

var (
	debug    = flag.Bool("debug", false, "enable debugging")
	password = flag.String("password", "", "the database password")
	port     = flag.Int("port", 1433, "the database port")
	server   = flag.String("server", "", "the database server")
	user     = flag.String("user", "", "the database user")
)

func main() {
	flag.Parse()
	if *debug {
		fmt.Printf(" password:%s\n", *password)
		fmt.Printf(" port:%d\n", *port)
		fmt.Printf(" server:%s\n", *server)
		fmt.Printf(" user:%s\n", *user)
	}

	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d", *server, *user, *password, *port)
	if *debug {
		fmt.Printf(" connString:%s\n", connString)
	}

	conn, err := sql.Open("mssql", connString)
	if err != nil {
		log.Fatal("Open connection failed:", err.Error())
	}
	defer conn.Close()

	stmt, err := conn.Prepare("select * from episodes")
	if err != nil {
		log.Fatal("Prepare failed:", err.Error())
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		log.Fatal("Query failed:", err.Error())
	}
	defer rows.Close()

	f, err := os.Create("episodes.txt")
	if err != nil {
		log.Fatal("Cannot create file:", err.Error())
	}
	defer f.Close()

	for rows.Next() {

		data := map[string]interface{}{
			"Id":                nil,
			"Number":            nil,
			"Title":             nil,
			"Description":       nil,
			"DateRecorded":      nil,
			"AudioFilePath":     nil,
			"AudioFileLength":   nil,
			"NumberOfDownloads": nil,
			"ZippedFilePath":    nil,
		}

		var myid int
		var number sql.NullInt64
		var title sql.NullString
		var description sql.NullString
		var dateRecorded sql.NullString
		var audioFilePath sql.NullString
		var audiofilelength sql.NullInt64
		var numberofdownloads sql.NullInt64
		var zipped sql.NullString
		err = rows.Scan(&myid, &number, &title, &description, &dateRecorded, &audioFilePath, &audiofilelength, &numberofdownloads, &zipped)
		if err != nil {
			log.Fatal("Scan failed:", err.Error())
		}

		data["id"] = myid
		data["Title"], _ = title.Value()
		data["Number"], _ = number.Value()
		data["Description"], _ = description.Value()
		data["DateRecorded"], _ = dateRecorded.Value()
		data["AudioFilePath"], _ = audioFilePath.Value()
		data["AudioFileLength"], _ = audiofilelength.Value()
		data["NumberOfDownloads"], _ = numberofdownloads.Value()
		data["ZippedFilePath"], _ = zipped.Value()

		t := template.Must(template.New("episode").Parse(episodeTemplate))
		buf := &bytes.Buffer{}
		if err := t.Execute(buf, data); err != nil {
			log.Fatal("Error executing the template", err.Error())
		}

		fmt.Print(buf)

		val, _ := data["Title"]
		fmt.Printf("my id :%s\n", val)

	}

}

const episodeTemplate = `+++
title = {{.Title}}
audio_file = {{ .AudioFilePath }}
date = {{ .Date }}
audio_length = {{ .AudioFileLength }}
guests = xxxxx
number = {{.Number}}
+++
{{ .Description }}
`
