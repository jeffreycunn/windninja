// Copyright 2017 Kyle Shannon.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func trueValue(s string) bool {
	switch s {
	case "true", "yes", "on":
		return true
	}
	return false
}

const reportSQL = `
SELECT country, region, COUNT()
	FROM visit LEFT JOIN ip USING(ip)
	GROUP BY region
	ORDER BY COUNT() DESC`

func reportHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	rows, err := db.Query(reportSQL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var g geogData
	var gs []geogData
	for rows.Next() {
		// TODO(kyle): check for errors, null values pass through
		rows.Scan(
			&g.Country,
			&g.Region,
			&g.Visits)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		gs = append(gs, g)
	}
	if r.FormValue("fmt") == "json" {
		var j []byte
		if trueValue(r.FormValue("pretty")) {
			j, err = json.MarshalIndent(gs, "", "  ")
		} else {
			j, err = json.Marshal(gs)
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		fmt.Fprintf(w, string(j))
	} else {
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")
		err = templates.ExecuteTemplate(w, "geogreport", gs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

const userSQL = `
SELECT ip, x, y, COUNT() as c
	FROM visit JOIN ip USING(ip)
	GROUP BY ip ORDER BY c DESC`

func userReport(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	type user struct {
		IP string  `json:"ip"`
		X  float64 `json:"x"`
		Y  float64 `json:"y"`
		C  int     `json:"count"`
	}
	var u user
	var us []user
	rows, err := db.Query(userSQL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&u.IP, &u.X, &u.Y, &u.C)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		us = append(us, u)
	}
	var j []byte
	if trueValue(r.FormValue("pretty")) {
		j, err = json.MarshalIndent(us, "", "  ")
	} else {
		j, err = json.Marshal(us)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	fmt.Fprintf(w, string(j))
}
