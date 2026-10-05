package researchpacket

import "errors"

const maxPacket = 10

// PromoteWithinBaseline moves the priority winner only when baseline already
// packed it. The baseline support set, packet cap, and other relative order
// remain unchanged; this is a ranking experiment, not a retrieval expansion.
func PromoteWithinBaseline(baseline, priority []string) ([]string, error) {
	if len(baseline) > maxPacket || len(priority) > maxPacket {
		return nil, errors.New("packet exceeds declared cap")
	}
	seen := make(map[string]bool, len(baseline))
	for _, id := range baseline {
		if id == "" || seen[id] {
			return nil, errors.New("baseline packet has empty or duplicate ID")
		}
		seen[id] = true
	}
	seenPriority := make(map[string]bool, len(priority))
	for _, id := range priority {
		if id == "" || seenPriority[id] {
			return nil, errors.New("priority packet has empty or duplicate ID")
		}
		seenPriority[id] = true
	}
	out := append([]string(nil), baseline...)
	if len(priority) == 0 || len(out) == 0 || !seen[priority[0]] {
		return out, nil
	}
	for i, id := range out {
		if id == priority[0] {
			copy(out[1:i+1], out[:i])
			out[0] = id
			return out, nil
		}
	}
	panic("validated priority winner disappeared from baseline")
}
