package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func makeSat(name string, position float32) Satellite {
	sat := Satellite{}
	_ = sat.SetName(name)
	sat.SetPosition(position)
	return sat
}

func TestFindNew(t *testing.T) {
	alien := makeSat("three", 3)
	list1 := []Satellite{makeSat("one", 1), makeSat("two", 2)}
	list2 := []Satellite{makeSat("one", 1), alien, makeSat("two", 2)}

	newItems := FindNewElements(&list1, &list2)

	assert.Len(t, newItems, 1)
	assert.Equal(t, newItems[0], alien)
}

func TestFindAbsence(t *testing.T) {
	alien := makeSat("three", 3)
	list1 := []Satellite{makeSat("one", 1), alien, makeSat("two", 2)}
	list2 := []Satellite{makeSat("one", 1), makeSat("two", 2)}

	absenceItems := FindAbsent(&list1, &list2)

	assert.Len(t, absenceItems, 1)
	assert.Equal(t, absenceItems[0], alien)
}

func TestChangedPosition(t *testing.T) {
	initial := makeSat("three", 3)
	changed := initial
	changed.SetPosition(4)
	list1 := []Satellite{makeSat("one", 1), initial, makeSat("two", 2)}
	list2 := []Satellite{changed, makeSat("one", 1), makeSat("two", 2)}

	absenceItems := FindChanged(&list1, &list2)

	assert.Len(t, absenceItems, 1)
	assert.Equal(t, absenceItems[0][0], initial)
	assert.Equal(t, absenceItems[0][1], changed)
}

func TestChangedURL(t *testing.T) {
	initial := makeSat("three", 3)
	changed := initial
	err := changed.SetURL(baseURL + "SAT.html")
	if assert.NoError(t, err) {
		list1 := []Satellite{makeSat("one", 1), initial, makeSat("two", 2)}
		list2 := []Satellite{changed, makeSat("one", 1), makeSat("two", 2)}

		absenceItems := FindChanged(&list1, &list2)

		assert.Len(t, absenceItems, 1)
		assert.Equal(t, absenceItems[0][0], initial)
		assert.Equal(t, absenceItems[0][1], changed)
	}
}

func TestChangedBand(t *testing.T) {
	initial := makeSat("three", 3)
	changed := initial
	changed.SetBand("band")
	list1 := []Satellite{makeSat("one", 1), initial, makeSat("two", 2)}
	list2 := []Satellite{changed, makeSat("one", 1), makeSat("two", 2)}

	absenceItems := FindChanged(&list1, &list2)

	assert.Len(t, absenceItems, 1)
	assert.Equal(t, absenceItems[0][0], initial)
	assert.Equal(t, absenceItems[0][1], changed)
}
