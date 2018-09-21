package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	BASE_ESPNFF_URL = "http://games.espn.com/ffl/api/v2"
	LOEG_LEAGUE_ID  = 365177
	//ALUM_LEAGUE_ID  = 1010746 // is a private league
	SEASON = 2018
)

func (l *league) New(id int) (*league, error) {

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

	// If the league is known, populate the team
	teams := []*team{}
	for _, leagueId := range knownLeagues {
		if leagueId == id {
			for _, t := range s.Teams {
				for _, o := range t.Owners {
					sb.Reset()
					sb.WriteString(o.FirstName)
					sb.WriteString(" ")
					sb.WriteString(o.LastName[:1])
					teams = append(teams, &team{
						Owner:         &owner{sb.String()},
						Abbreviation:  t.TeamAbbrev,
						Id:            t.TeamId,
						Wins:          t.Record.OverallWins,
						PointsFor:     t.Record.PointsFor,
						PointsAgainst: t.Record.PointsAgainst,
					})
				}
			}
		}
	}
	return &league{
		Id:    id,
		Teams: teams,
	}, nil
}

type league struct {
	Id    int
	Teams []*team
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

var (
	wins = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "wins",
			Help: "Number of wins on the season.",
		},
		[]string{"owner"})

	pointsFor = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pointsFor",
			Help: "Points scored this season",
		},
		[]string{"owner"})

	pointsAgainst = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pointsAgainst",
			Help: "Points against this season",
		},
		[]string{"owner"})

	knownLeagues = []int{LOEG_LEAGUE_ID} // , ALUM_LEAGUE_ID}
)

func init() {
	prometheus.MustRegister(wins)
	prometheus.MustRegister(pointsFor)
	prometheus.MustRegister(pointsAgainst)
}

func main() {

	var loeg league
	loegLeague, err := loeg.New(LOEG_LEAGUE_ID)
	if err != nil {
		log.Fatal(err)
	}

	for _, t := range loegLeague.Teams {
		name := t.Owner.Name
		wins.With(prometheus.Labels{"owner": name}).Set(t.Wins)
		pointsFor.With(prometheus.Labels{"owner": name}).Set(t.PointsFor)
		pointsAgainst.With(prometheus.Labels{"owner": name}).Set(t.PointsAgainst)
	}

	log.Println("Starting FF metrics server at /ff-metrics...")
	http.Handle("/ff-metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":9090", nil))
}
