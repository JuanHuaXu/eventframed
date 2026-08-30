package libravdbstore

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	libra "github.com/xDarkicex/libravdb/libravdb"
)

// MigrateLegacy backs up first, copies into model-keyed collections, and
// publishes schema state last. It deliberately leaves legacy collections intact.
func MigrateLegacy(ctx context.Context, config Config, activeEmbedder embed.Embedder, backupPath string) error {
	if !filepath.IsAbs(backupPath) {
		return errors.New("migration backup path must be absolute")
	}
	if config.EmbeddingModel == "" || config.Dimension <= 0 {
		return errors.New("migration requires an embedding model and positive dimension")
	}
	if activeEmbedder == nil || activeEmbedder.ModelKey() != config.EmbeddingModel || activeEmbedder.Dimension() != config.Dimension {
		return errors.New("migration embedder does not match the destination embedding contract")
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
			if raw, ok := record.Metadata["raw_content"].(string); ok && raw != "" {
				event.Content = raw
			}
			vector, embedErr := embed.Document(activeEmbedder, event.FrameText())
			if embedErr != nil {
				return fmt.Errorf("embed legacy EventFrame %q: %w", event.ID, embedErr)
			}
			event.Embedding = nil
			event.EmbeddingModel = config.EmbeddingModel
			destinationName := collectionName(event.TenantID, config.EmbeddingModel)
			if _, err := db.EnsureCollection(ctx, destinationName, config.Dimension, migrationCollectionOptions(config)...); err != nil {
				return err
			}
			eventJSON, _ := encodeStoredEvent(event)
			metadata := cloneMetadata(record.Metadata)
			metadata["event_json"] = string(eventJSON)
			metadata["corpus_text"] = event.FrameText()
			metadata["raw_content"] = event.Content
			if priorDigest, ok := metadata["content_digest"].(string); !ok || priorDigest == "" {
				digest := sha256.Sum256([]byte(event.Content))
				metadata["content_digest"] = hex.EncodeToString(digest[:])
			}
			if err := db.WithTx(ctx, func(tx libra.Tx) error { return tx.Upsert(ctx, destinationName, event.ID, vector, metadata) }); err != nil {
				return err
			}
			count++
		}
	}
	snapshot := model.Snapshot{RuntimeVersion: count, EvidenceEpoch: count, PolicyVersion: 1, ContractVersion: model.ContractVersion, GraphVersion: 1, PosteriorVersion: 1, ResidualVersion: 1, AbstractionVersion: 1, AgencyVersion: 1}
	metadata, err := (&Store{config: config}).stateMetadata(snapshot)
	if err != nil {
		return err
	}
	if err := system.Insert(ctx, "runtime", nil, metadata); err != nil {
		return fmt.Errorf("publish migrated state: %w", err)
	}
	return nil
}

// MigrateEventFrameCorpus re-embeds a durable pre-v12 database from raw-text
// vectors into canonical 5W1H vectors. The source collections remain intact.
func MigrateEventFrameCorpus(ctx context.Context, config Config, activeEmbedder embed.Embedder, backupPath string) error {
	if !filepath.IsAbs(backupPath) {
		return errors.New("migration backup path must be absolute")
	}
	if activeEmbedder == nil || activeEmbedder.ModelKey() != config.EmbeddingModel || activeEmbedder.Dimension() != config.Dimension {
		return errors.New("migration embedder does not match the destination embedding contract")
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
	record, err := system.Get(ctx, "runtime")
	if err != nil {
		return fmt.Errorf("read durable runtime state: %w", err)
	}
	encoded, ok := record.Metadata["state_json"].(string)
	if !ok {
		return errors.New("durable runtime state is malformed")
	}
	var state persistentState
	if err := json.Unmarshal([]byte(encoded), &state); err != nil {
		return fmt.Errorf("decode durable runtime state: %w", err)
	}
	if state.SchemaVersion != 2 || state.Dimension != config.Dimension || state.Quantization != config.Quantization {
		return errors.New("source database does not match the requested durable embedding shape")
	}
	if state.EmbeddingModel != embed.UnbindRepresentation(config.EmbeddingModel) {
		return fmt.Errorf("source embedding contract %q is not the predecessor of %q", state.EmbeddingModel, config.EmbeddingModel)
	}
	names, err := db.ListCollectionsWithContext(ctx)
	if err != nil {
		return err
	}
	count := uint64(0)
	for _, sourceName := range names {
		if sourceName == systemCollection || sourceName == bayesianCollection || sourceName == agencyCollection {
			continue
		}
		source, getErr := db.GetCollection(sourceName)
		if getErr != nil {
			continue
		}
		records, listErr := source.ListAll(ctx)
		if listErr != nil {
			return listErr
		}
		for _, sourceRecord := range records {
			payload, hasEvent := sourceRecord.Metadata["event_json"].(string)
			if !hasEvent {
				continue
			}
			var event model.Event
			if err := json.Unmarshal([]byte(payload), &event); err != nil {
				return fmt.Errorf("decode event %q during corpus migration: %w", sourceRecord.ID, err)
			}
			if raw, ok := sourceRecord.Metadata["raw_content"].(string); ok && raw != "" {
				event.Content = raw
			}
			vector, embedErr := embed.Document(activeEmbedder, event.FrameText())
			if embedErr != nil {
				return fmt.Errorf("embed EventFrame %q: %w", event.ID, embedErr)
			}
			event.Embedding = nil
			event.EmbeddingModel = config.EmbeddingModel
			destinationName := collectionName(event.TenantID, config.EmbeddingModel)
			if _, err := db.EnsureCollection(ctx, destinationName, config.Dimension, migrationCollectionOptions(config)...); err != nil {
				return err
			}
			eventJSON, err := encodeStoredEvent(event)
			if err != nil {
				return err
			}
			metadata := cloneMetadata(sourceRecord.Metadata)
			metadata["event_json"] = string(eventJSON)
			metadata["corpus_text"] = event.FrameText()
			metadata["raw_content"] = event.Content
			if err := db.WithTx(ctx, func(tx libra.Tx) error { return tx.Upsert(ctx, destinationName, event.ID, vector, metadata) }); err != nil {
				return err
			}
			count++
		}
	}
	if count == 0 {
		return errors.New("no durable events found to migrate")
	}
	state.EmbeddingModel = config.EmbeddingModel
	state.Snapshot.ContractVersion = model.ContractVersion
	state.Snapshot.RuntimeVersion++
	state.Snapshot.PolicyVersion++
	state.Snapshot.AbstractionVersion++
	migrationStore := &Store{config: config, policyDigest: state.BayesianPolicyDigest}
	metadata, err := migrationStore.stateMetadata(state.Snapshot)
	if err != nil {
		return err
	}
	if err := system.Upsert(ctx, "runtime", nil, metadata); err != nil {
		return fmt.Errorf("publish EventFrame corpus state: %w", err)
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
