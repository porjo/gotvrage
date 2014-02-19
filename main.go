package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

func main() {

	type Show struct {
		Name    string `xml:"name,attr"`
		Sid     int    `xml:"sid"`
		Network string `xml:"network"`
		Title   string `xml:"title"`
		Episode string `xml:"ep"`
		Link    string `xml:"link"`
	}
	type Hour struct {
		TimeStr   string `xml:"attr,attr"`
		Shows     []Show `xml:"show"`
		Timestamp time.Time
	}
	type Day struct {
		DayDate string `xml:"attr,attr"`
		Hours   []Hour `xml:"time"`
	}
	type Query struct {
		Days []Day `xml:"DAY"`
	}

	res, err := http.Get("http://services.tvrage.com/feeds/fullschedule.php?country=AU")
	if err != nil {
		fmt.Printf("Get:", err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	v := Query{}

	err = xml.Unmarshal(body, &v)
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}

	vtmp := make(map[string]map[string][]Hour)

	for _, day := range v.Days {
		for _, hour := range day.Hours {
			timestamp, err := time.Parse("2006-1-2 03:04 pm -0700", day.DayDate+" "+hour.TimeStr+" +1000")
			if err != nil {
				fmt.Printf("error: %v", err)
				return
			}
			hour.Timestamp = timestamp

			yearMonth := timestamp.Format("2006-01")
			dayStr := timestamp.Format("02")

			d, ok := vtmp[yearMonth]
			if !ok {
				d = make(map[string][]Hour)
				vtmp[yearMonth] = d
			}

			d[dayStr] = append(d[dayStr], hour)
		}
	}

	for yearMonth := range vtmp {

		b, err := json.Marshal(vtmp[yearMonth])
		if err != nil {
			fmt.Println(err)
			return
		}

		if _, err := os.Stat(yearMonth); os.IsNotExist(err) {
			err = os.Mkdir(yearMonth, 0750)
			if err != nil {
				fmt.Printf("error: %v", err)
				return
			}
		}
		err = ioutil.WriteFile(yearMonth+"/data.json", b, 0640)
		if err != nil {
			fmt.Printf("error: %v", err)
			return
		}
	}
}
