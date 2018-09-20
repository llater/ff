package ff

import (

    "strings"
    "strconv"
    "net/http"
    "io/ioutil"
    "log"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "encoding/json"
)

const (
    BASE_ESPNFF_URL = "http://games.espn.com/ffl/api/v2"
    LOEG_LEAGUE_ID = 365177
    ALUM_LEAGUE_ID = 1010746
    SEASON = 2018
)

type league struct {
    Id int
    Teams []*team
}

type standings struct {
    Teams []struct {
        TeamAbbrev string `json:"teamAbbrev"`
        Owners []struct {
            FirstName string `json:"firstName"`
            LastName string `json:"lastName"`
            Id int `json:"teamId"`
        }
        Record struct {
            OverallWins int `json:"overallWins"`
            PointsFor float32 `json:"pointsFor"`
            PointsAgainst float32 `json:"pointsAgainst"`
        } `json:"record"`


    } `json:"teams"`
}

func (* league) MustCreate(id int) *league {
    // Look up the league on ESPN
    sb := strings.Builder{}
    sb.WriteString(BASE_ESPNFF_URL)
    sb.WriteString("/standings?leagueId=")
    sb.WriteString(strconv.Itoa(id))
    sb.WriteString("&seasonId=")
    sb.WriteString(strconv.Itoa(SEASON))

    resp, err := http.Get(sb.String())
    if err != nil { log.Fatal(err) }
    b, err := ioutil.ReadAll(resp.Body)
    defer resp.Body.Close()
    if err != nil { log.Fatal(err) }
    var s standings
    if err := json.Unmarshal(b, &s); err != nil { log.Fatal(err) }
}

type team struct {
    Owner *owner
    Abbreviation string
    Id int
}

type owner struct {
    Name string
}



var (
    wins = prometheus.NewCounterVec(
        prometheus.CounterOpts{
        Name: "wins",
        Help: "Number of wins on the season.",
    },
        []string{"owner"})

    loegMembersMap = map[int]*owner{
        12: &owner{ "Leland"},
}

alumMembersMap = map[int]*owner{

}

)



func init() {
    prometheus.MustRegister(wins)

}

func main() {
    qsb := strings.Builder{}
    qsb.WriteString(BASE_ESPNFF_URL)
    qsb.WriteString("/standings")
    qsb.WriteString("?leagueId=")
    qsb.WriteString(strconv.Itoa())


}