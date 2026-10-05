package researchindex

import (
	"context"
	"errors"
	"slices"
)

// IncomingRemoval is a private deletion STAGE. It removes incoming links but
// does not reconnect neighbors, retire the target, or choose an entry point.
// These edits are not a complete deletion transaction and must not be published
// alone as one. Every returned record/slice is owned by the caller.
type IncomingRemoval struct {
	Edits       []LayeredEdit
	Affected    [][]uint32
	ReadRecords int
}

func DiscoverIncomingRemoval(ctx context.Context, before LayeredSnapshot, target uint32, maxReads, maxEdits int) (IncomingRemoval, error) {
	return discoverIncomingLevel(ctx, before, target, maxReads, maxEdits, -1)
}
func discoverIncomingLevel(ctx context.Context, before LayeredSnapshot, target uint32, maxReads, maxEdits, onlyLevel int) (IncomingRemoval, error) {
	var out IncomingRemoval
	if maxReads < 1 || maxReads > 65536 || maxEdits < 1 || maxEdits > 128 {
		return out, ErrCapacity
	}
	seen := map[uint32]bool{}
	read := func(id uint32) (*LayeredRecord, error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if !seen[id] {
			if len(seen) >= maxReads {
				return nil, ErrCapacity
			}
			seen[id] = true
		}
		return layeredFind(before.tree, id), nil
	}
	t, err := read(target)
	if err != nil {
		return IncomingRemoval{}, err
	}
	if t == nil {
		return IncomingRemoval{}, errors.New("incoming target absent")
	}
	out.Affected = make([][]uint32, t.Level+1)
	edited := map[uint32]*LayeredRecord{}
	for level := 0; level <= t.Level; level++ {
		if onlyLevel >= 0 && level != onlyLevel {
			continue
		}
		for _, id := range t.Backlinks[level] {
			n, err := read(id)
			if err != nil {
				return IncomingRemoval{}, err
			}
			if n == nil || n.Level < level {
				continue
			}
			if pending := edited[id]; pending != nil {
				n = pending
			}
			at := slices.Index(n.Links[level], target)
			if at < 0 {
				continue
			}
			owned := edited[id]
			if owned == nil {
				if len(edited) >= maxEdits {
					return IncomingRemoval{}, ErrCapacity
				}
				copy, ok := before.Lookup(id)
				if !ok {
					return IncomingRemoval{}, errors.New("immutable record disappeared")
				}
				owned = &copy
				edited[id] = owned
			}
			links := owned.Links[level]
			links[at] = links[len(links)-1]
			owned.Links[level] = links[:len(links)-1]
			out.Affected[level] = append(out.Affected[level], id)
		}
	}
	for id, r := range edited {
		out.Edits = append(out.Edits, LayeredEdit{id, r})
	}
	slices.SortFunc(out.Edits, func(a, b LayeredEdit) int {
		if a.Ordinal < b.Ordinal {
			return -1
		}
		if a.Ordinal > b.Ordinal {
			return 1
		}
		return 0
	})
	if err := ctx.Err(); err != nil {
		return IncomingRemoval{}, err
	}
	out.ReadRecords = len(seen)
	return out, nil
}
