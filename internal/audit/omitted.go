package audit

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"math"
	"sort"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

type Estimate struct {
	PopulationSize int
	SelectedIDs    []string
	Mean           float64
	UpperBound     float64
}

func Selected(eventID, seed string, epoch uint64, probability float64) bool {
	if probability >= 1 {
		return true
	}
	if probability <= 0 {
		return false
	}
	payload := seed + "\x00" + eventID + "\x00" + string(binary.BigEndian.AppendUint64(nil, epoch))
	digest := sha256.Sum256([]byte(payload))
	draw := float64(binary.BigEndian.Uint64(digest[:8])) / float64(^uint64(0))
	return draw < probability
}

func EstimateInfluence(population []string, local model.BernoulliLaw, expanded map[string]model.BernoulliLaw, seed string, epoch, sequence uint64, probability, delta float64) (Estimate, error) {
	if len(population) == 0 || probability <= 0 || probability > 1 || delta <= 0 || delta >= 1 || sequence == 0 {
		return Estimate{}, errors.New("audit requires a finite population, nonzero inclusion probability, sequence, and confidence delta")
	}
	if !validLaw(local) {
		return Estimate{}, errors.New("local law is invalid")
	}
	seen := make(map[string]struct{}, len(population))
	selected := make([]string, 0)
	sum := 0.0
	for _, eventID := range population {
		if eventID == "" {
			return Estimate{}, errors.New("audit population ids cannot be empty")
		}
		if _, duplicate := seen[eventID]; duplicate {
			return Estimate{}, errors.New("audit population ids must be unique")
		}
		seen[eventID] = struct{}{}
		if !Selected(eventID, seed, epoch, probability) {
			continue
		}
		law, ok := expanded[eventID]
		if !ok || !validLaw(law) {
			return Estimate{}, errors.New("every selected audit event requires a valid expanded law")
		}
		selected = append(selected, eventID)
		sum += JensenShannon(local, law) / probability
	}
	for eventID := range expanded {
		if _, ok := seen[eventID]; !ok || !Selected(eventID, seed, epoch, probability) {
			return Estimate{}, errors.New("expanded laws must match the preselected audit sample")
		}
	}
	if len(selected) == 0 {
		return Estimate{}, errors.New("audit sample is empty")
	}
	sort.Strings(selected)
	mean := sum / float64(len(population))
	spentDelta := delta / (float64(sequence) * float64(sequence+1))
	radius := math.Sqrt(math.Log(1/spentDelta)/(2*float64(len(population)))) / probability
	return Estimate{PopulationSize: len(population), SelectedIDs: selected, Mean: mean, UpperBound: math.Min(1, mean+radius)}, nil
}

func JensenShannon(left, right model.BernoulliLaw) float64 {
	middleUseful := (left.Useful + right.Useful) / 2
	middleNotUseful := (left.NotUseful + right.NotUseful) / 2
	value := (kl(left.Useful, middleUseful) + kl(left.NotUseful, middleNotUseful) + kl(right.Useful, middleUseful) + kl(right.NotUseful, middleNotUseful)) / (2 * math.Log(2))
	return math.Max(0, math.Min(1, value))
}

func kl(value, reference float64) float64 {
	if value == 0 {
		return 0
	}
	return value * math.Log(value/reference)
}

func validLaw(law model.BernoulliLaw) bool {
	return law.Useful >= 0 && law.Useful <= 1 && law.NotUseful >= 0 && law.NotUseful <= 1 && math.Abs(law.Useful+law.NotUseful-1) <= 1e-9
}
