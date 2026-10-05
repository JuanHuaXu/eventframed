// eventframe-research-baseline reports the prospective lab reference; it does not configure the daemon.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/JuanHuaXu/eventframed/internal/researchbaseline"
)

func main() {
	registry := flag.String("registry", "research/baselines.json", "prospective research reference registry")
	flag.Parse()
	b, e := os.ReadFile(*registry)
	if e != nil {
		fail(e)
	}
	var r struct {
		Cohort     string                       `json:"cohort"`
		Primary    string                       `json:"primary"`
		Candidates []researchbaseline.Candidate `json:"candidates"`
	}
	if e = json.Unmarshal(b, &r); e != nil {
		fail(e)
	}
	chosen, e := researchbaseline.Select(r.Candidates, r.Cohort)
	if e != nil {
		fail(e)
	}
	if chosen.Name != r.Primary || r.Primary != researchbaseline.PrimaryArm {
		fail(fmt.Errorf("registry and reviewed prospective policy disagree"))
	}
	if e = json.NewEncoder(os.Stdout).Encode(struct {
		Scope             string                     `json:"scope"`
		Primary           researchbaseline.Candidate `json:"primary"`
		SpeedControl      string                     `json:"speed_control"`
		ProductionChanged bool                       `json:"production_changed"`
	}{"Prospective consumed-cohort research reference only", chosen, researchbaseline.SpeedControl, false}); e != nil {
		fail(e)
	}
}

func fail(e error) { fmt.Fprintln(os.Stderr, e); os.Exit(1) }
