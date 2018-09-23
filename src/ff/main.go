package main

import (
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	knownLeagues = []int{LOEG_LEAGUE_ID} // , ALUM_LEAGUE_ID}
)

const (
	BASE_ESPNFF_URL = "http://games.espn.com/ffl/api/v2"
	LOEG_LEAGUE_ID  = 365177
	//ALUM_LEAGUE_ID  = 1010746 // is a private league
	SEASON = 2018
)

type league struct {
	Id         int
	Teams      []*team
	IdOwnerMap map[int]string
}

func getStandings(id int) (*standings, error) {
	// Build a query string from disparate components
	sb := strings.Builder{}
	sb.WriteString(BASE_ESPNFF_URL)
	sb.WriteString("/standings?leagueId=")
	sb.WriteString(strconv.Itoa(id))
	sb.WriteString("&seasonId=")
	sb.WriteString(strconv.Itoa(SEASON))

	resp, err := http.Get(sb.String())
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return nil, errors.New("private league")
	}

	decoder := json.NewDecoder(resp.Body)

	var s standings
	if err := decoder.Decode(&s); err != nil {
		return nil, err
	}

	return &s,nil
}

func (l *league) New(id int) (*league, error) {

	s, err := getStandings(id)
	if err != nil {
		return nil, err
	}

	sb := strings.Builder{}
	// If the league is known, populate the teams
	teams := []*team{}
	idOwnerName := make(map[int]string)
	for _, leagueId := range knownLeagues {
		if leagueId == id {
			for _, t := range s.Teams {
				for _, owners := range t.Owners {
					sb.Reset()
					sb.WriteString(owners.FirstName)
					sb.WriteString(" ")
					sb.WriteString(owners.LastName[:1])
					ownerName := sb.String()
					teams = append(teams, &team{
						Owner:         &owner{Name: ownerName},
						Abbreviation:  t.TeamAbbrev,
						Id:            t.TeamId,
						Wins:          t.Record.OverallWins,
						PointsFor:     t.Record.PointsFor,
						PointsAgainst: t.Record.PointsAgainst,
					})
					idOwnerName[t.TeamId] = ownerName
				}
			}
		}
	}
	return &league{
		Id:         id,
		Teams:      teams,
		IdOwnerMap: idOwnerName,
	}, nil
}

type team struct {
	Owner         *owner
	Abbreviation  string
	Id            int
	Wins          float64
	PointsFor     float64
	PointsAgainst float64
}

type owner struct {
	Name string
}

type standings struct {
	Teams []struct {
		TeamAbbrev string `json:"teamAbbrev"`
		Owners     []struct {
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
			Id        int    `json:"teamId"`
		} `json:"owners"`
		Record struct {
			OverallWins   float64 `json:"overallWins"`
			PointsFor     float64 `json:"pointsFor"`
			PointsAgainst float64 `json:"pointsAgainst"`
		} `json:"record"`
		TeamId int `json:"teamId"`
	} `json:"teams"`
}

func init() {
	prometheus.MustRegister(wins)
	prometheus.MustRegister(pointsForTotal)
	prometheus.MustRegister(pointsAgainstTotal)
	prometheus.MustRegister(pointsByWeek)
}

func main() {

	var l league
	loeg, err := l.New(LOEG_LEAGUE_ID)
	if err != nil {
		log.Fatal(err)
	}

	weekly := flag.Bool("weekly",false,"Scrape these metrics once a week.")
	gametime := flag.Bool("gametime", false, "Scrape these metrics twice a minute during games.")

	flag.Parse()

	if *weekly {
		loeg.CollectWeekly()
	}
	if *gametime {
		loeg.CollectLiveGames()
	}

	log.Println("Starting FF metrics server at /metrics...")
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":8080", nil))
}
