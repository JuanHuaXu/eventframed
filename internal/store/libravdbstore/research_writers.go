package libravdbstore

// OpenResearchFourWriters is an opt-in storage experiment. It raises the pinned
// backend's per-collection execution bound from two to four, retaining queue
// depth64, synchronous WAL acknowledgment and all store snapshot/identity locks.
// No async indexing or unsafe durability mode is enabled. Open remains control;
// no daemon configuration selects this constructor.
func OpenResearchFourWriters(config Config) (*Store, error) {
	return openWithWriteConcurrency(config, 4)
}
