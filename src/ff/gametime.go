package main

import (
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var (
	pointsByWeek = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pointsWeek",
			Help: "Points scored on a given week",
		},
		[]string{"owner", "week"})
)

type scoreboard struct {
	Scoreboard struct {
		Matchups []struct {
			Teams []struct {
					Score  float64 `json:"score"`
					TeamId int     `json:"teamId"`
			} `json:"teams"`
		} `json:"matchups"`
		Week int `json:"matchupPeriodId"`
	} `json:"scoreboard"`

}

func (l *league) CollectLiveGames() {
	// Collect the scoreboard
	sb := strings.Builder{}
	sb.WriteString(BASE_ESPNFF_URL)
	sb.WriteString("/scoreboard?leagueId=")
	sb.WriteString(strconv.Itoa(l.Id))
	sb.WriteString("&seasonId=")
	sb.WriteString(strconv.Itoa(SEASON))

	resp, err := http.Get(sb.String())
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		log.Println("private league")
	}

	decoder := json.NewDecoder(resp.Body)

	var s scoreboard
	if err := decoder.Decode(&s); err != nil {
		log.Fatal(err)
	}

	for _, matchup := range s.Scoreboard.Matchups {
		for _, team := range matchup.Teams {
			ownerName := l.IdOwnerMap[team.TeamId]
			pointsByWeek.With(prometheus.Labels{"owner": ownerName, "week": strconv.Itoa(s.Scoreboard.Week)}).Set(team.Score)
		}
	}
}

