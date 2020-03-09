package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/html/charset"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const satellitePositionPattern string = "^(?i)([0-9.]+)°([EW])$"

var (
	baseURL                = getProperties().Parser.BaseURL
	sourceUrls             = []string{baseURL + "asia.html", baseURL + "america.html", baseURL + "atlantic.html", baseURL + "europe.html"}
	satellitePositionRegex = regexp.MustCompile(satellitePositionPattern)
)

func lastLevelTable(_ int, selection *goquery.Selection) bool {
	return selection.Has("table").Length() == 0
}

func containsVerdana(_ int, selection *goquery.Selection) bool {
	outerHTML, err := selection.Html()
	return err == nil && strings.Contains(outerHTML, "Verdana")
}

// Parse extracts satellite items from given reader and sends them to given chData channel.
// Occurred errors are sent to chErr channel. url string is used for tracing purposes only.
func Parse(url string, reader io.Reader, chData chan Satellite, chErr chan error) {
	log.Infof("parsing started: %s", url)

	document, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		err = fmt.Errorf("error reading HTTP response body: %w", err)
		log.Error(err)
		chErr <- err
		return
	}

	satellite := Satellite{}
	gotPosition := false
	doneCounter := 0

	allCounter := document.
		Find("table").
		FilterFunction(lastLevelTable).
		FilterFunction(containsVerdana).
		Children().Unwrap().
		Find("tr").
		Each(func(_ int, selection *goquery.Selection) {
			data, _ := selection.Html()

			length := selection.Children().Length()
			if length != 4 && length != 5 {
				err := fmt.Errorf("wrong format, there must be 4 or 5 tds, but got %d. Data: %s", length, data)
				log.Error(err)
				chErr <- err
				return
			}

			if length == 5 {
				positionString := selection.Children().Eq(1).Text()

				matches := satellitePositionRegex.FindAllStringSubmatch(positionString, -1)

				if matches == nil {
					err := fmt.Errorf("position string doesn't match regex '%s'. Data: %s", satellitePositionPattern, data)
					log.Error(err)
					chErr <- err
					return
				}

				position, err := strconv.ParseFloat(matches[0][1], 64)
				if err != nil {
					err := fmt.Errorf("cannot parse satellite position: %w. Data: %s", err, data)
					log.Error(err)
					chErr <- err
					return
				}

				if strings.EqualFold(matches[0][2], "W") {
					position *= -1
				}

				satellite.SetPosition(float32(position))
				gotPosition = true
			} else if !gotPosition {
				err := fmt.Errorf("satellite doesn't have position neither previous satellite. Data: %s", data)
				log.Error(err)
				chErr <- err
				return
			}

			nameTd := selection.Children().Eq(length - 3)

			if err := satellite.SetName(nameTd.Text()); err != nil {
				err := fmt.Errorf("satellite's name setting error: %w. Data: %s", err, data)
				log.Error(err)
				chErr <- err
				return
			}

			a := nameTd.Find("a")
			url, exists := a.Attr("href")
			if !exists {
				err := fmt.Errorf("cannot find satellite url. Data: %s", data)
				log.Error(err)
				chErr <- err
				return
			}

			if err1 := satellite.SetURL(url); err1 != nil {
				err2 := fmt.Errorf("%w. Name: %s, url: %s", err1, satellite.Name, url)
				chErr <- err2
				return
			}

			satellite.SetBand(selection.Children().Eq(length - 2).Text())

			log.Debugf("satellite parsed: %v", satellite)
			chData <- satellite
			doneCounter++
		}).Length()
	log.Infof("parsing finished: %s. %d out of %d satellites processed", url, doneCounter, allCounter)
}

func getResponse(url string) (*http.Response, error) {
	log.Printf("loading content of %s ...", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	log.Printf("got response from %s", url)

	if resp.StatusCode != 200 {
		err = fmt.Errorf("cannot get document (%s): status code is %d", url, resp.StatusCode)
		return nil, err
	}

	return resp, nil
}

func closeReader(response *http.Response) error {
	return response.Body.Close()
}

func getUtf8Reader(response *http.Response) (io.Reader, error) {
	contentType := response.Header.Get("Content-Type")
	reader, err := charset.NewReader(response.Body, contentType)
	if err != nil {
		return nil, fmt.Errorf("cannot convert document to utf-8: %w", err)
	}

	return reader, nil
}

// parseOnline runs pages parsing in goroutines, compiles, sorts and returns satellites array.
func parseOnline() ([]Satellite, []error) {
	ch, chErr, chQuit := make(chan Satellite), make(chan error), make(chan int)
	ongoing := 0

	for _, url := range sourceUrls {
		go parseOnlinePage(url, ch, chErr, chQuit)
	}

	var satellites []Satellite
	var errorz []error
WaiterLoop:
	for {
		select {
		case receivedSat := <-ch:
			satellites = append(satellites, receivedSat)
		case receivedErr := <-chErr:
			errorz = append(errorz, receivedErr)
		case count := <-chQuit:
			ongoing += count
			if ongoing == 0 {
				break WaiterLoop
			}
		}
	}
	close(ch)
	close(chErr)
	close(chQuit)

	sort.Sort(ByPosName(satellites))

	return satellites, errorz
}

func parseOnlinePage(url string, chData chan Satellite, chErr chan error, chCounter chan int) {
	chCounter <- 1
	defer func() {
		chCounter <- -1
	}()

	httpResponse, err := getResponse(url)
	if err != nil {
		chErr <- err
		return
	}
	defer func() {
		if err := closeReader(httpResponse); err != nil {
			chErr <- err
		}
	}()

	reader, err := getUtf8Reader(httpResponse)
	if err != nil {
		chErr <- err
		return
	}

	Parse(url, reader, chData, chErr)
}
