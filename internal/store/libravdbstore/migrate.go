package libravdbstore

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/JuanHuaXu/eventframed/internal/model"
	libra "github.com/xDarkicex/libravdb/libravdb"
)

// MigrateLegacy backs up first, copies into model-keyed collections, and
// publishes schema state last. It deliberately leaves legacy collections intact.
func MigrateLegacy(ctx context.Context, config Config, backupPath string) error {
	if !filepath.IsAbs(backupPath) {
		return errors.New("migration backup path must be absolute")
	}
	if config.EmbeddingModel == "" || config.Dimension <= 0 {
		return errors.New("migration requires an embedding model and positive dimension")
	}
	db, err := libra.Open(libra.WithStoragePath(config.Path), libra.WithMaxConcurrentWrites(1))
	if err != nil {
		return err
	}
	defer db.Close()
	if err := db.Backup(ctx, backupPath); err != nil {
		return fmt.Errorf("pre-migration backup: %w", err)
	}
	system, err := db.EnsureCollection(ctx, systemCollection, 0, libra.WithMetadataOnly())
	if err != nil {
		return err
	}
	if _, err := system.Get(ctx, "runtime"); err == nil {
		return errors.New("database already has durable schema state")
	} else if !errors.Is(err, libra.ErrRecordNotFound) {
		return fmt.Errorf("inspect durable schema state: %w", err)
	}
	names, err := db.ListCollectionsWithContext(ctx)
	if err != nil {
		return err
	}
	legacy := make([]string, 0)
	for _, name := range names {
		if isLegacyCollection(name) {
			legacy = append(legacy, name)
		}
	}
	if len(legacy) == 0 {
		return errors.New("no legacy Phase 1 collections found")
	}
	count := uint64(0)
	for _, sourceName := range legacy {
		source, err := db.GetCollection(sourceName)
		if err != nil {
			return err
		}
		records, err := source.ListAll(ctx)
		if err != nil {
			return err
		}
		for _, record := range records {
			payload, ok := record.Metadata["event_json"].(string)
			if !ok {
				continue
			}
			var event model.Event
			if json.Unmarshal([]byte(payload), &event) != nil {
				continue
			}
			event.Embedding = nil
			event.EmbeddingModel = config.EmbeddingModel
			destinationName := collectionName(event.TenantID, config.EmbeddingModel)
			if _, err := db.EnsureCollection(ctx, destinationName, config.Dimension, migrationCollectionOptions(config)...); err != nil {
				return err
			}
			eventJSON, _ := json.Marshal(event)
			digest := sha256.Sum256(eventJSON)
			metadata := cloneMetadata(record.Metadata)
			metadata["event_json"] = string(eventJSON)
			metadata["content_digest"] = hex.EncodeToString(digest[:])
			if err := db.WithTx(ctx, func(tx libra.Tx) error { return tx.Upsert(ctx, destinationName, event.ID, record.Vector, metadata) }); err != nil {
				return err
			}
			count++
		}
	}
	snapshot := model.Snapshot{RuntimeVersion: count, EvidenceEpoch: count, PolicyVersion: 1, ContractVersion: model.ContractVersion, GraphVersion: 1, PosteriorVersion: 1, ResidualVersion: 1, AbstractionVersion: 1}
	metadata, err := (&Store{config: config}).stateMetadata(snapshot)
	if err != nil {
		return err
	}
	if err := system.Insert(ctx, "runtime", nil, metadata); err != nil {
		return fmt.Errorf("publish migrated state: %w", err)
	}
	return nil
}

func migrationCollectionOptions(config Config) []libra.CollectionOption {
	options := []libra.CollectionOption{libra.WithMetric(libra.CosineDistance), libra.WithHNSW(16, 200, 100), libra.WithMemoryMapping(config.MemoryMapping)}
	switch config.Quantization {
	case "sq8":
		options = append(options, libra.WithScalarQuantization(8, 0.10))
	case "fsq6":
		options = append(options, libra.WithFSQQuantization(6, 0.10))
	case "pq8":
		options = append(options, libra.WithProductQuantization(8, 8, 0.10))
	}
	return options
}

func cloneMetadata(source map[string]interface{}) map[string]interface{} {
	copy := make(map[string]interface{}, len(source))
	for key, value := range source {
		copy[key] = value
	}
	return copy
}
