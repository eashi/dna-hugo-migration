package main

import (
	"bytes"
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"io"
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
	panic("Open connection failed:", err)
	defer conn.Close()

	stmt, err := conn.Prepare("select * from episodes")
	panic("Prepare failed:", err)
	defer stmt.Close()

	rows, err := stmt.Query()
	panic("Query failed:", err)
	defer rows.Close()

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

		var f, err = os.Create(fmt.Sprintf("%d.md", data["Number"]))
		panic("couldn't create the md file", err)
		defer f.Close()

		t := template.Must(template.New("episode").Parse(episodeTemplate))
		buf := &bytes.Buffer{}
		executeOrPanic(t.Execute, buf, data, "Error executing the template")

		fmt.Print(buf)
		f.WriteString(buf.String())
	}

}

const episodeTemplate = `+++
title = {{.Title}}
audio_file = {{ .AudioFilePath }}
date = {{ .DateRecorded }}
audio_length = {{ .AudioFileLength }}
guests = xxxxx
number = {{.Number}}
+++
{{ .Description }}
`

func panic(message string, err error) {
	if err != nil {
		log.Fatal(message, err.Error())
	}
}

func executeOrPanic(fn func(io.Writer, interface{}) error, arg1 io.Writer, arg2 interface{}, message string) {
	err := fn(arg1, arg2)
	if err != nil {
		log.Fatal(message, err.Error())
	}
}
