package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql" // mysql driver is used explicitly in sqlx
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
)

var (
	dbPtr                     *sqlx.DB
	selectActiveStmt          = "SELECT _position, _name, _url, _band FROM `%s` WHERE _status = 1 ORDER BY _position, _name"
	insertSatelliteStmt       = "INSERT INTO `%s` (_name, _position, _url, _band, _tags) VALUES (:_name, :_position, :_url, :_band, '')"
	updateSatelliteStatusStmt = "UPDATE `%s` SET _status = 0, _closed = CURRENT_TIMESTAMP() WHERE _name = :_name AND _status != 0"
	updateSatelliteStmt       = "UPDATE `%s` SET _position = :_position, _url = :_url, _band = :_band WHERE _name = :_name AND _status = 1"
)

func updateProperties() {
	selectActiveStmt = fmt.Sprintf(selectActiveStmt, getProperties().Mysql.Table)
	insertSatelliteStmt = fmt.Sprintf(insertSatelliteStmt, getProperties().Mysql.Table)
	updateSatelliteStatusStmt = fmt.Sprintf(updateSatelliteStatusStmt, getProperties().Mysql.Table)
	updateSatelliteStmt = fmt.Sprintf(updateSatelliteStmt, getProperties().Mysql.Table)
}

func getDB() *sqlx.DB {
	if dbPtr == nil {
		log.Info("opening connection to MySQL ...")
		updateProperties()
		var err error
		dbPtr, err = sqlx.Open("mysql", getProperties().Mysql.URL)
		if err != nil {
			log.WithError(err).Fatal("critical error, shutting down ...")
		}
		log.Info("connection to MySQL created")
	}
	return dbPtr
}

// LoadDbSatellites loads all active satellite items from database.
func LoadDbSatellites() []Satellite {
	log.Info("loading satellites from MySQL ...")

	var satellites []Satellite
	if err := getDB().Select(&satellites, selectActiveStmt); err != nil {
		log.WithError(err).Fatal("critical error, shutting down ...")
	}

	log.Infof("DB loading finished. %d satellites loaded", len(satellites))

	return satellites
}

func InsertSatellites(list *[]Satellite) {
	log.Info("inserting satellites to MySQL ...")

	count := 0
	if len(*list) > 0 {
		tx := getDB().MustBegin()
		for _, sat := range *list {
			log.Debugf("inserting %v", sat)
			_, err := tx.NamedExec(insertSatelliteStmt, sat)
			if err != nil {
				log.WithError(err).Errorf("cannot insert satellite %v", sat)
				continue
			}
			count++
		}
		err := tx.Commit()
		if err != nil {
			log.WithError(err).Error("cannot insert satellites, transaction error")
		}
	}
	log.Infof("inserting satellites finished. %d out of %d inserted", len(*list), count)
}

func MarkSatellitesClosed(list *[]Satellite) {
	log.Info("marking satellites closed in MySQL ...")

	count := 0
	if len(*list) > 0 {
		tx := getDB().MustBegin()
		for _, sat := range *list {
			log.Debugf("marking %v", sat)
			_, err := tx.NamedExec(updateSatelliteStatusStmt, sat)
			if err != nil {
				log.WithError(err).Errorf("cannot mark satellite %v", sat)
				continue
			}
			count++
		}
		err := tx.Commit()
		if err != nil {
			log.WithError(err).Error("cannot mark satellites, transaction error")
		}
	}
	log.Infof("marking satellites closed finished. %d out of %d marked", len(*list), count)
}

func UpdateSatellites(list *[][]Satellite) {
	log.Info("updating satellites in MySQL ...")

	count := 0
	if len(*list) > 0 {
		tx := getDB().MustBegin()
		for _, pair := range *list {
			oldSat, newSat := pair[0], pair[1]
			log.Debugf("updating %v with new values %v", oldSat, newSat)
			_, err := tx.NamedExec(updateSatelliteStmt, newSat)
			if err != nil {
				log.WithError(err).Errorf("cannot update satellite %v with new values %v", oldSat, newSat)
				continue
			}
			count++
		}
		err := tx.Commit()
		if err != nil {
			log.WithError(err).Error("cannot update satellites, transaction error")
		}
	}
	log.Infof("updating satellites finished. %d out of %d marked", len(*list), count)
}
