package main

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/jmoiron/sqlx"

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

	db, err := sqlx.Connect("mssql", connString)
	panic("no!", err)

	episodes := []Episode{}
	err2 := db.Select(&episodes, "SELECT * FROM Episodes")
	panic("selectx episodes failed", err2)

	guests := []Guest{}
	errGuests := db.Select(&guests, "SELECT * FROM Guests")
	panic("selectx guests failed", errGuests)

	episodeGuests := []EpisodeGuest{}
	errepisodeGuests := db.Select(&episodeGuests, "SELECT * FROM EpisodesGuests")
	panic("selectx guests failed", errepisodeGuests)

	defer db.Close()

	if *debug {
		fmt.Printf(" connString:%s\n", connString)
	}

	funcMap := template.FuncMap{
		"valueOf": valueOf,
	}
	episodeTemplateInstance := template.Must(template.New("episode").Funcs(funcMap).Parse(episodeTemplate))
	guestTemplateInstance := template.Must(template.New("guest").Funcs(funcMap).Parse(guestTemplate))

	for index := range guests {

		var thisGuest = &guests[index]
		guestImagePath, _ := thisGuest.ImagePath.Value()
		guestImagePathTemp := strings.ReplaceAll(guestImagePath.(string), "\\", "/")
		guestImagePathTemp = strings.ToLower(guestImagePathTemp)
		guestImagePathTemp = strings.ReplaceAll(guestImagePathTemp, "images/guests/", "")
		guestImagePathTemp = strings.ReplaceAll(guestImagePathTemp, ".jpg", "")
		guestImagePathTemp = strings.ReplaceAll(guestImagePathTemp, ".gif", "")

		guests[index].EnglishName = guestImagePathTemp

		//create the file guest
		var f, err = os.Create(fmt.Sprintf("%s.md", (*thisGuest).EnglishName))
		panic("couldn't create the guest md file", err)
		defer f.Close()

		buf := &bytes.Buffer{}
		executeOrPanic(guestTemplateInstance.Execute, buf, thisGuest, "Error executing the Guest template")
		f.WriteString(buf.String())

		// fmt.Println(buf)

	}

	for index := range episodes {

		tempEpisode := &episodes[index]

		// guestsIds := make([]string, 10)
		guestsIds := []string{}
		audioFilePath, _ := (*tempEpisode).AudioFilePath.Value()
		audioFilePath = strings.ReplaceAll(audioFilePath.(string), "\\", "/")

		quotesEscapedTitle, _ := (*tempEpisode).Title.Value()
		(*tempEpisode).Title = sql.NullString{String: strings.ReplaceAll(quotesEscapedTitle.(string), "\"", "\\\""), Valid: true}

		(*tempEpisode).AudioFilePath = sql.NullString{String: strings.ReplaceAll(audioFilePath.(string), "audio/", ""), Valid: true}

		for _, episodeguest := range episodeGuests {
			if episodeguest.EpisodeID == (*tempEpisode).ID {
				for _, tempGuest := range guests {
					if tempGuest.ID == episodeguest.GuestID {
						guestsIds = append(guestsIds, fmt.Sprintf("\"%s\"", tempGuest.EnglishName))
					}
				}
			}
		}

		(*tempEpisode).Guests = strings.Join(guestsIds, ",")

		numberValue, _ := (*tempEpisode).Number.Value()
		var f, err3 = os.Create(fmt.Sprintf("%d.md", numberValue))
		panic("couldn't create the md file", err3)
		defer f.Close()

		buf := &bytes.Buffer{}
		executeOrPanic(episodeTemplateInstance.Execute, buf, *tempEpisode, "Error executing the template")
		f.WriteString(buf.String())
	}

}

const episodeTemplate = `+++
title = "{{ valueOf .Title }}"
audio_file = "{{ valueOf .AudioFilePath }}"
date = {{ valueOf .DateRecorded }}
audio_length = "{{ valueOf .AudioFileLength }}"
guests = [{{ .Guests }}]
number = {{valueOf .Number}}
+++
{{ valueOf .Description }}
`

const guestTemplate = `+++
Title = "{{ valueOf .FullName }}"
image = "{{ .EnglishName }}.jpg"
+++
{{ valueOf .Description }}
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

type Episode struct {
	ID                sql.NullInt64  `db:"Id"`
	Number            sql.NullInt64  `db:"Number"`
	Title             sql.NullString `db:"Title"`
	Description       sql.NullString `db:"Description"`
	DateRecorded      sql.NullString `db:"DateRecorded"`
	AudioFileLength   sql.NullInt64  `db:"AudioFileLength"`
	AudioFilePath     sql.NullString `db:"AudioFilePath"`
	ZippedFilePath    sql.NullString `db:"ZippedFilePath"`
	NumberOfDownloads sql.NullInt64  `db:"NumberOfDownloads"`
	Guests            string
}

type Guest struct {
	ID          sql.NullInt64  `db:"Id"`
	FullName    sql.NullString `db:"FullName"`
	Description sql.NullString `db:"Description"`
	ImagePath   sql.NullString `db:"ImagePath"`
	EnglishName string
}

type EpisodeGuest struct {
	EpisodeID sql.NullInt64 `db:"EpisodeId"`
	GuestID   sql.NullInt64 `db:"GuestId"`
}

func valueOf(arg sqlDbType) driver.Value {
	value, _ := arg.Value()
	if value == nil {
		return ""
	}

	return value

}

type sqlDbType interface {
	Value() (driver.Value, error)
}
