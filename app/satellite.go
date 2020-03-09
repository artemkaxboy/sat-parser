package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"regexp"
	"strings"
)

// Satellite is a struct to hold all information about satellites.
type Satellite struct {
	Name     string  `db:"_name"`
	URL      string  `db:"_url"`
	Position float32 `db:"_position"`
	Band     string  `db:"_band"`
}

var (
	satelliteNameTailRegex = regexp.MustCompile(`\W*\(.*$`)
	satelliteURLPattern    = getProperties().Parser.SatelliteURLPattern
	satelliteURLRegex      = regexp.MustCompile(satelliteURLPattern)

	relativeURLRegex = regexp.MustCompile(`^[^:]*$`)
)

// SetName removes additional information in brackets from satellite name e.g. (incl 0.6), trims spaces
// and sets the value into struct.
// Returns error if the value is empty.
func (ptr *Satellite) SetName(name string) error {
	name = satelliteNameTailRegex.ReplaceAllString(name, "") // removes additional information from name (e.g. incl 0.6)
	name = strings.TrimSpace(name)
	if len(name) > 0 {
		ptr.Name = name
		return nil
	}

	err := fmt.Errorf("satellite name cannot be empty")
	log.WithError(err).Debug(err)
	return err
}

// GetName returns name field as is
func (ptr *Satellite) GetName() string {
	return ptr.Name
}

func isURLCorrect(url string) bool {
	return satelliteURLRegex.Match([]byte(url))
}

func isURLRelative(url string) bool {
	return relativeURLRegex.Match([]byte(url))
}

// SetURL checks that url matches allowed Regexp and sets it.
// Returns error if the value does not match allowed Regexp.
func (ptr *Satellite) SetURL(url string) error {
	if isURLCorrect(url) {
		ptr.URL = url
		return nil
	}

	if isURLRelative(url) {
		ptr.URL = getProperties().Parser.BaseURL + url
		return nil
	}

	err := fmt.Errorf("satellite url does not match expected regex '%s'", satelliteURLPattern)
	log.WithError(err).Debugf("cannot set url %s", url)
	return err
}

// GetURL returns url field as is.
func (ptr *Satellite) GetURL() string {
	return ptr.URL
}

// SetPosition sets position field as is.
func (ptr *Satellite) SetPosition(position float32) {
	ptr.Position = position
}

// GetPosition returns position field as is.
func (ptr *Satellite) GetPosition() float32 {
	return ptr.Position
}

// SetBand trims the value and sets it.
func (ptr *Satellite) SetBand(band string) {
	ptr.Band = strings.TrimSpace(band)
}

// GetBand returns band field as is.
func (ptr *Satellite) GetBand() string {
	return ptr.Band
}

// ByPosName is utility type to sort Satellites array.
type ByPosName []Satellite

func (a ByPosName) Len() int { return len(a) }
func (a ByPosName) Less(i, j int) bool {
	iPosition := a[i].Position
	jPosition := a[j].Position
	if iPosition != jPosition {
		return iPosition < jPosition
	}
	return a[i].Name < a[j].Name
}
func (a ByPosName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
