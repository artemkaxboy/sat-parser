package main

import (
	log "github.com/sirupsen/logrus"
)

func getOnlineList(ch chan []Satellite) {
	onlineList, errorz := parseOnline()

	log.Infof("online parsing finished, satellites count - %d", len(onlineList))

	if errorzLen := len(errorz); errorzLen > 0 {
		for _, err := range errorz {
			log.Error(err)
		}

		log.Fatalf("some errors [%d] occurred during parsing, check them at first", errorzLen)
	}

	ch <- onlineList
}

func getDBList(ch chan []Satellite) {
	ch <- LoadDbSatellites()
}

func getLists() (onlineList []Satellite, dbList []Satellite) {
	chOnline, chDB := make(chan []Satellite), make(chan []Satellite)
	defer func() {
		close(chOnline)
		close(chDB)
	}()

	go getOnlineList(chOnline)
	go getDBList(chDB)

	return <-chOnline, <-chDB
}

func main() {
	level, err := log.ParseLevel(getProperties().LogLevel)
	if err == nil {
		log.SetLevel(level)
	}

	onlineList, dbList := getLists()

	newItems := FindNewElements(&dbList, &onlineList)
	if len(newItems) > 0 {
		InsertSatellites(&newItems)
	}

	absentItems := FindAbsent(&dbList, &onlineList)
	if len(absentItems) > 0 {
		MarkSatellitesClosed(&absentItems)
	}

	changedItems := FindChanged(&dbList, &onlineList)
	if len(changedItems) > 0 {
		UpdateSatellites(&changedItems)
	}
}
