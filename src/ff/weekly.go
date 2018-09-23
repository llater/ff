package main

import (
	"github.com/prometheus/client_golang/prometheus"

"log"
)
var (
	wins = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "wins",
			Help: "Number of wins on the season.",
		},
		[]string{"owner"})

	pointsForTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pointsForTotal",
			Help: "Points scored this season",
		},
		[]string{"owner"})

	pointsAgainstTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pointsAgainstTotal",
			Help: "Points against this season",
		},
		[]string{"owner"})
)


func (l *league) CollectWeekly() {
	s, err := getStandings(l.Id)
	if err != nil {
		log.Fatal(err)
	}
	for _, team := range s.Teams {
        ownerName := l.IdOwnerMap[team.TeamId]
        wins.With(prometheus.Labels{"owner":ownerName}).Set(team.Record.OverallWins)
        pointsForTotal.With(prometheus.Labels{"owner":ownerName}).Set(team.Record.PointsFor)
        pointsAgainstTotal.With(prometheus.Labels{"owner":ownerName}).Set(team.Record.PointsAgainst)
	}

}
