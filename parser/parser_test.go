package main

import (
	"github.com/stretchr/testify/assert"
	"io"
	"sort"
	"strings"
	"testing"
)

func collect(testName string, sample string) ([]Satellite, []error) {
	ch, chErr, chQuit := make(chan Satellite), make(chan error), make(chan int)
	ongoing := 0

	go collectPage(testName, strings.NewReader(sample), ch, chErr, chQuit)

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

func collectPage(testName string, reader io.Reader, chData chan Satellite, chErr chan error, chCounter chan int) {
	chCounter <- 1
	defer func() {
		chCounter <- -1
	}()

	Parse(testName, reader, chData, chErr)
}

func TestTableLevel(t *testing.T) {
	sample := `<table cellspacing=0 border>
<tr>
<td bgcolor="#ffbf00" width=1><font size=2>&nbsp;</font></td><td width=70 rowspan=2 bgcolor=khaki align="center"><font face="Verdana"><font size=2><a href="` + getProperties().Parser.BaseURL + `Optus-D3-10.html">156.</font><font size=1>0</font><font size=2>&#176;E</font></a></td>
<td width=180 bgcolor=khaki><font face="Arial"><font size=2><a href="` + getProperties().Parser.BaseURL + `Optus-D3.html">Optus D3</a></td>
<td width=20 bgcolor=khaki><font face="Arial"><font size=1></font><font size=1> Ku</font></td>
<td width=50 bgcolor=#ffffff align=center><font face="Verdana" size=1>181103</td>
</tr>
<tr><td><table/></td></tr>
</table>`

	satellites, errs := collect(t.Name(), sample)

	assert.Empty(t, satellites, "should not find table because of another inner table")
	assert.Empty(t, errs)
}

func TestLackOfVerdana(t *testing.T) {
	sample := `<table cellspacing=0 border>
<tr>
<td bgcolor="#ffbf00" width=1><font size=2>&nbsp;</font></td><td width=70 rowspan=2 bgcolor=khaki align="center"><font face="Arial"><font size=2><a href="` + getProperties().Parser.BaseURL + `Optus-D3-10.html">156.</font><font size=1>0</font><font size=2>&#176;E</font></a></td>
<td width=180 bgcolor=khaki><font face="Arial"><font size=2><a href="` + getProperties().Parser.BaseURL + `Optus-D3.html">Optus D3</a></td>
<td width=20 bgcolor=khaki><font face="Arial"><font size=1></font><font size=1> Ku</font></td>
<td width=50 bgcolor=#ffffff align=center><font face="Arial" size=1>181103</td>
</tr>
</table>`
	satellites, errs := collect(t.Name(), sample)

	assert.Empty(t, satellites)
	assert.Empty(t, errs)
}

func TestTooFewTdCount(t *testing.T) {
	sample := `<table cellspacing=0 border>
<tr>
<td bgcolor="#ffbf00" width=1><font size=2>&nbsp;</font></td><td width=70 rowspan=2 bgcolor=khaki align="center"><font face="Verdana"><font size=2><a href="` + getProperties().Parser.BaseURL + `Optus-D3-10.html">156.</font><font size=1>0</font><font size=2>&#176;E</font></a></td>
<td width=180 bgcolor=khaki><font face="Arial"><font size=2><a href="` + getProperties().Parser.BaseURL + `Optus-D3.html">Optus D3</a></td>
</tr>
</table>`
	satellites, errs := collect(t.Name(), sample)

	assert.Empty(t, satellites)
	assert.Len(t, errs, 1)
}

func TestTooManyTdCount(t *testing.T) {
	sample := `<table cellspacing=0 border>
<tr>
<td bgcolor="#ffbf00" width=1><font size=2>&nbsp;</font></td><td width=70 rowspan=2 bgcolor=khaki align="center"><font face="Verdana"><font size=2><a href="` + getProperties().Parser.BaseURL + `Optus-D3-10.html">156.</font><font size=1>0</font><font size=2>&#176;E</font></a></td>
<td width=180 bgcolor=khaki><font face="Arial"><font size=2><a href="` + getProperties().Parser.BaseURL + `Optus-D3.html">Optus D3</a></td>
<td width=20 bgcolor=khaki><font face="Arial"><font size=1></font><font size=1> Ku</font></td>
<td width=50 bgcolor=#ffffff align=center><font face="Verdana" size=1>181103</td>
<td></td>
</tr>
</table>`

	satellites, errs := collect(t.Name(), sample)
	assert.Empty(t, satellites)
	assert.Len(t, errs, 1)
}

func TestRowspan(t *testing.T) {
	sample := `<table cellspacing=0 border>
<tr>
<td bgcolor="#ffbf00" width=1><font size=2>&nbsp;</font></td><td width=70 rowspan=2 bgcolor=khaki align="center"><font face="Verdana"><font size=2><a href="` + getProperties().Parser.BaseURL + `Optus-D3-10.html">156.</font><font size=1>0</font><font size=2>&#176;E</font></a></td>
<td width=180 bgcolor=khaki><font face="Arial"><font size=2><a href="` + getProperties().Parser.BaseURL + `Optus-D3.html">Optus D3</a></td>
<td width=20 bgcolor=khaki><font face="Arial"><font size=1></font><font size=1> Ku</font></td>
<td width=50 bgcolor=#ffffff align=center><font face="Verdana" size=1>181103</td>
</tr>
<tr>
<td bgcolor="white" width=1><font size=2>&nbsp;</font></td><td width=180 bgcolor=khaki><font face="Arial"><font size=2><a href="` + getProperties().Parser.BaseURL + `Optus-10.html">Optus 10</a></td>
<td width=20 bgcolor=khaki><font face="Arial"><font size=1></font><font size=1> Ku</font></td>
<td width=50 bgcolor=#ffffff align=center><font face="Verdana" size=1>190520</td>
</tr>
</table>`
	satellites, errs := collect(t.Name(), sample)

	assert.Len(t, satellites, 2)
	assert.Empty(t, errs)
	assert.Equal(t, satellites[0].GetPosition(), satellites[1].GetPosition(), "should have the same position")
}

func TestIncl(t *testing.T) {
	sample := `<table cellspacing=0 border>
<tr>
<td bgcolor="white" width=1><font size=2>&nbsp;</font></td><td width=70 rowspan=3 bgcolor=khaki align="center"><font face="Verdana"><font size=2><a href="` + getProperties().Parser.BaseURL + `ABS-7-and-Koreasat-6-7.html">116.</font><font size=1>0</font><font size=2>&#176;E</font></a></td>
<td width=180 bgcolor=khaki><font face="Arial"><font size=2><a href="` + getProperties().Parser.BaseURL + `ABS-7.html">ABS 7</a> <i><font face="Arial" size=1><a href="` + getProperties().Parser.BaseURL + `tracker/ABS-7.html">(incl. 0.<font size=1>6</font>°)</a></i></td>
<td width=20 bgcolor=khaki><font face="Arial"><font size=1></font><font size=1> Ku</font></td>
<td width=50 bgcolor=#ffffff align=center><font face="Verdana" size=1>120507</td>
</tr>
</table>`
	satellites, errs := collect(t.Name(), sample)

	assert.Len(t, satellites, 1)
	assert.Empty(t, errs)

	assert.Equal(t, "ABS 7", satellites[0].GetName())
	assert.Equal(t, getProperties().Parser.BaseURL+"ABS-7.html", satellites[0].GetURL())
}

func TestLackOfPosition(t *testing.T) {
	sample := `<table cellspacing=0 border>
<tr>
<td bgcolor="white" width=1><font size=2>&nbsp;</font></td>
<td width=180 bgcolor=khaki><font face="Arial"><font size=2><a href="` + getProperties().Parser.BaseURL + `ABS-7.html">ABS 7</a> <i><font face="Arial" size=1><a href="` + getProperties().Parser.BaseURL + `tracker/ABS-7.html">(incl. 0.<font size=1>6</font>°)</a></i></td>
<td width=20 bgcolor=khaki><font face="Arial"><font size=1></font><font size=1> Ku</font></td>
<td width=50 bgcolor=#ffffff align=center><font face="Verdana" size=1>120507</td>
</tr>
</table>`

	satellites, errs := collect(t.Name(), sample)

	assert.Empty(t, satellites)
	assert.Len(t, errs, 1)
}

func TestBandParsing(t *testing.T) {
	sample := `<table cellspacing=0 border>
<tr>
<td bgcolor="white" width=1><font size=2>&nbsp;</font></td><td width=70 rowspan=3 bgcolor=khaki align="center"><font face="Verdana"><font size=2><a href="` + getProperties().Parser.BaseURL + `ABS-7-and-Koreasat-6-7.html">116.</font><font size=1>0</font><font size=2>&#176;E</font></a></td>
<td width=180 bgcolor=khaki><font face="Arial"><font size=2><a href="` + getProperties().Parser.BaseURL + `ABS-7.html">ABS 7</a> <i><font face="Arial" size=1><a href="` + getProperties().Parser.BaseURL + `tracker/ABS-7.html">(incl. 0.<font size=1>6</font>°)</a></i></td>
<td width=20 bgcolor=khaki><font face="Arial"><font size=1></font><font size=1> Ku</font></td>
<td width=50 bgcolor=#ffffff align=center><font face="Verdana" size=1>120507</td>
</tr>
</table>`
	satellites, errs := collect(t.Name(), sample)

	assert.Len(t, satellites, 1)
	assert.Empty(t, errs)

	assert.Equal(t, "Ku", satellites[0].GetBand())
}

func TestWrongUrl(t *testing.T) {
	sample := `<table cellspacing=0 border>
<tr>
<td bgcolor="white" width=1><font size=2>&nbsp;</font></td><td width=70 rowspan=3 bgcolor=khaki align="center"><font face="Verdana"><font size=2><a href="` + getProperties().Parser.BaseURL + `ABS-7-and-Koreasat-6-7.html">116.</font><font size=1>0</font><font size=2>&#176;E</font></a></td>
<td width=180 bgcolor=khaki><font face="Arial"><font size=2><a href="` + getProperties().Parser.BaseURL + `tracker/ABS-7.html">ABS 7</a> <i><font face="Arial" size=1><a href="` + getProperties().Parser.BaseURL + `tracker/ABS-7.html">(incl. 0.<font size=1>6</font>°)</a></i></td>
<td width=20 bgcolor=khaki><font face="Arial"><font size=1></font><font size=1> Ku</font></td>
<td width=50 bgcolor=#ffffff align=center><font face="Verdana" size=1>120507</td>
</tr>
</table>`
	satellites, errs := collect(t.Name(), sample)

	assert.Empty(t, satellites)
	assert.Len(t, errs, 1)
}

func TestLackOfUrl(t *testing.T) {
	sample := `<table cellspacing=0 border>
<tr>
<td bgcolor="white" width=1><font size=2>&nbsp;</font></td><td width=70 rowspan=3 bgcolor=khaki align="center"><font face="Verdana"><font size=2><a href="` + getProperties().Parser.BaseURL + `ABS-7-and-Koreasat-6-7.html">116.</font><font size=1>0</font><font size=2>&#176;E</font></a></td>
<td width=180 bgcolor=khaki><font face="Arial"><font size=2>ABS 7 <i><font face="Arial" size=1>(incl. 0.<font size=1>6</font>°)</i></td>
<td width=20 bgcolor=khaki><font face="Arial"><font size=1></font><font size=1> Ku</font></td>
<td width=50 bgcolor=#ffffff align=center><font face="Verdana" size=1>120507</td>
</tr>
</table>`
	satellites, errs := collect(t.Name(), sample)

	assert.Empty(t, satellites)
	assert.Len(t, errs, 1)
}

func TestWrongNumberPosition(t *testing.T) {
	sample := `<table cellspacing=0 border>
<tr>
<td bgcolor="white" width=1><font size=2>&nbsp;</font></td><td width=70 rowspan=3 bgcolor=khaki align="center"><font face="Verdana"><font size=2><a href="` + getProperties().Parser.BaseURL + `ABS-7-and-Koreasat-6-7.html">116..</font><font size=1>0</font><font size=2>&#176;E</font></a></td>
<td width=180 bgcolor=khaki><font face="Arial"><font size=2><a href="` + getProperties().Parser.BaseURL + `tracker/ABS-7.html">ABS 7</a> <i><font face="Arial" size=1><a href="` + getProperties().Parser.BaseURL + `tracker/ABS-7.html">(incl. 0.<font size=1>6</font>°)</a></i></td>
<td width=20 bgcolor=khaki><font face="Arial"><font size=1></font><font size=1> Ku</font></td>
<td width=50 bgcolor=#ffffff align=center><font face="Verdana" size=1>120507</td>
</tr>
</table>`
	satellites, errs := collect(t.Name(), sample)

	assert.Empty(t, satellites)
	assert.Len(t, errs, 1)
}

func TestWrongFormatPosition(t *testing.T) {
	sample := `<table cellspacing=0 border>
<tr>
<td bgcolor="white" width=1><font size=2>&nbsp;</font></td><td width=70 rowspan=3 bgcolor=khaki align="center"><font face="Verdana"><font size=2><a href="` + getProperties().Parser.BaseURL + `ABS-7-and-Koreasat-6-7.html">116.</font><font size=1>0</font><font size=2>&#176;U</font></a></td>
<td width=180 bgcolor=khaki><font face="Arial"><font size=2><a href="` + getProperties().Parser.BaseURL + `ABS-7.html">ABS 7</a> <i><font face="Arial" size=1><a href="` + getProperties().Parser.BaseURL + `tracker/ABS-7.html">(incl. 0.<font size=1>6</font>°)</a></i></td>
<td width=20 bgcolor=khaki><font face="Arial"><font size=1></font><font size=1> Ku</font></td>
<td width=50 bgcolor=#ffffff align=center><font face="Verdana" size=1>120507</td>
</tr>
</table>`
	satellites, errs := collect(t.Name(), sample)

	assert.Empty(t, satellites)
	assert.Len(t, errs, 1)
}

func TestEPosition(t *testing.T) {
	sample := `<table cellspacing=0 border>
<tr>
<td bgcolor="white" width=1><font size=2>&nbsp;</font></td><td width=70 rowspan=3 bgcolor=khaki align="center"><font face="Verdana"><font size=2><a href="` + getProperties().Parser.BaseURL + `ABS-7-and-Koreasat-6-7.html">116.</font><font size=1>0</font><font size=2>&#176;E</font></a></td>
<td width=180 bgcolor=khaki><font face="Arial"><font size=2><a href="` + getProperties().Parser.BaseURL + `ABS-7.html">ABS 7</a> <i><font face="Arial" size=1><a href="` + getProperties().Parser.BaseURL + `tracker/ABS-7.html">(incl. 0.<font size=1>6</font>°)</a></i></td>
<td width=20 bgcolor=khaki><font face="Arial"><font size=1></font><font size=1> Ku</font></td>
<td width=50 bgcolor=#ffffff align=center><font face="Verdana" size=1>120507</td>
</tr>
</table>`
	satellites, errs := collect(t.Name(), sample)

	assert.Len(t, satellites, 1)
	assert.Empty(t, errs)
	assert.Equal(t, float32(116), satellites[0].GetPosition())
}

func TestWPosition(t *testing.T) {
	sample := `<table cellspacing=0 border>
<tr>
<td bgcolor="white" width=1><font size=2>&nbsp;</font></td><td width=70 rowspan=3 bgcolor=khaki align="center"><font face="Verdana"><font size=2><a href="` + getProperties().Parser.BaseURL + `ABS-7-and-Koreasat-6-7.html">116.</font><font size=1>0</font><font size=2>&#176;w</font></a></td>
<td width=180 bgcolor=khaki><font face="Arial"><font size=2><a href="` + getProperties().Parser.BaseURL + `ABS-7.html">ABS 7</a> <i><font face="Arial" size=1><a href="` + getProperties().Parser.BaseURL + `tracker/ABS-7.html">(incl. 0.<font size=1>6</font>°)</a></i></td>
<td width=20 bgcolor=khaki><font face="Arial"><font size=1></font><font size=1> Ku</font></td>
<td width=50 bgcolor=#ffffff align=center><font face="Verdana" size=1>120507</td>
</tr>
</table>`
	satellites, errs := collect(t.Name(), sample)

	assert.Len(t, satellites, 1)
	assert.Empty(t, errs)
	assert.Equal(t, float32(-116), satellites[0].GetPosition())
}

func TestEmptyName(t *testing.T) {
	sample := `<table cellspacing=0 border>
<tr>
<td bgcolor="white" width=1><font size=2>&nbsp;</font></td><td width=70 rowspan=3 bgcolor=khaki align="center"><font face="Verdana"><font size=2><a href="` + getProperties().Parser.BaseURL + `ABS-7-and-Koreasat-6-7.html">116.</font><font size=1>0</font><font size=2>&#176;w</font></a></td>
<td width=180 bgcolor=khaki><font face="Arial"><font size=2><a href="` + getProperties().Parser.BaseURL + `ABS-7.html">     </a> <i><font face="Arial" size=1><a href="` + getProperties().Parser.BaseURL + `tracker/ABS-7.html">(incl. 0.<font size=1>6</font>°)</a></i></td>
<td width=20 bgcolor=khaki><font face="Arial"><font size=1></font><font size=1> Ku</font></td>
<td width=50 bgcolor=#ffffff align=center><font face="Verdana" size=1>120507</td>
</tr>
</table>`
	satellites, errs := collect(t.Name(), sample)

	assert.Empty(t, satellites)
	assert.Len(t, errs, 1)
}
