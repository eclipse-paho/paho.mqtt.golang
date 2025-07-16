/*
 * Copyright (c) 2021 IBM Corp and others.
 *
 * All rights reserved. This program and the accompanying materials
 * are made available under the terms of the Eclipse Public License v2.0
 * and Eclipse Distribution License v1.0 which accompany this distribution.
 *
 * The Eclipse Public License is available at
 *    https://www.eclipse.org/legal/epl-2.0/
 * and the Eclipse Distribution License is available at
 *   http://www.eclipse.org/org/documents/edl-v10.php.
 *
 * Contributors:
 *    Seth Hoenig
 *    Allan Stockdill-Mander
 *    Mike Robertson
 */

package mqtt

import (
	"log/slog"
	"sync"

	"github.com/eclipse/paho.mqtt.golang/packets"
)

// MemoryStore implements the store interface to provide a "persistence"
// mechanism wholly stored in memory. This is only useful for
// as long as the client instance exists.
type MemoryStore struct {
	sync.RWMutex
	messages map[string]packets.ControlPacket
	opened   bool
	logger   *slog.Logger
}

// NewMemoryStore returns a pointer to a new instance of
// MemoryStore, the instance is not initialized and ready to
// use until Open() has been called on it.
func NewMemoryStore() *MemoryStore {
	store := &MemoryStore{
		messages: make(map[string]packets.ControlPacket),
		opened:   false,
		logger:   noopSLogger,
	}
	return store
}

// NewMemoryStoreEx returns a pointer to a new instance of MemoryStore with a custom logger.
func NewMemoryStoreEx(logger *slog.Logger) *MemoryStore {
	if logger == nil {
		logger = noopSLogger
	}
	store := &MemoryStore{
		messages: make(map[string]packets.ControlPacket),
		opened:   false,
		logger:   logger,
	}
	return store
}

// Open initializes a MemoryStore instance.
func (store *MemoryStore) Open() {
	store.Lock()
	defer store.Unlock()
	store.opened = true
	DEBUG.Println(STR, "memorystore initialized")
	store.logger.Debug("memorystore initialized", componentAttr(STR))
}

// Put takes a key and a pointer to a Message and stores the
// message.
func (store *MemoryStore) Put(key string, message packets.ControlPacket) {
	store.Lock()
	defer store.Unlock()
	if !store.opened {
		ERROR.Println(STR, "Trying to use memory store, but not open")
		store.logger.Error("Trying to use memory store, but not open", componentAttr(STR))
		return
	}
	store.messages[key] = message
}

// Get takes a key and looks in the store for a matching Message
// returning either the Message pointer or nil.
func (store *MemoryStore) Get(key string) packets.ControlPacket {
	store.RLock()
	defer store.RUnlock()
	if !store.opened {
		ERROR.Println(STR, "Trying to use memory store, but not open")
		store.logger.Error("Trying to use memory store, but not open", componentAttr(STR))
		return nil
	}
	mid := mIDFromKey(key)
	m := store.messages[key]
	if m == nil {
		CRITICAL.Println(STR, "memorystore get: message", mid, "not found")
		store.logger.Error("memorystore get: message not found", slog.Uint64("messageID", uint64(mid)), componentAttr(STR))
	} else {
		DEBUG.Println(STR, "memorystore get: message", mid, "found")
		store.logger.Debug("memorystore get: message found", slog.Uint64("messageID", uint64(mid)), componentAttr(STR))
	}
	return m
}

// All returns a slice of strings containing all the keys currently
// in the MemoryStore.
func (store *MemoryStore) All() []string {
	store.RLock()
	defer store.RUnlock()
	if !store.opened {
		ERROR.Println(STR, "Trying to use memory store, but not open")
		store.logger.Error("Trying to use memory store, but not open", componentAttr(STR))
		return nil
	}
	var keys []string
	for k := range store.messages {
		keys = append(keys, k)
	}
	return keys
}

// Del takes a key, searches the MemoryStore and if the key is found
// deletes the Message pointer associated with it.
func (store *MemoryStore) Del(key string) {
	store.Lock()
	defer store.Unlock()
	if !store.opened {
		ERROR.Println(STR, "Trying to use memory store, but not open")
		store.logger.Error("Trying to use memory store, but not open", componentAttr(STR))
		return
	}
	mid := mIDFromKey(key)
	m := store.messages[key]
	if m == nil {
		WARN.Println(STR, "memorystore del: message", mid, "not found")
		store.logger.Warn("memorystore del: message not found", slog.Uint64("messageID", uint64(mid)), componentAttr(STR))
	} else {
		delete(store.messages, key)
		DEBUG.Println(STR, "memorystore del: message", mid, "was deleted")
		store.logger.Debug("memorystore del: message was deleted", slog.Uint64("messageID", uint64(mid)), componentAttr(STR))
	}
}

// Close will disallow modifications to the state of the store.
func (store *MemoryStore) Close() {
	store.Lock()
	defer store.Unlock()
	if !store.opened {
		ERROR.Println(STR, "Trying to close memory store, but not open")
		store.logger.Error("Trying to close memory store, but not open", componentAttr(STR))
		return
	}
	store.opened = false
	DEBUG.Println(STR, "memorystore closed")
	store.logger.Debug("memorystore closed", componentAttr(STR))
}

// Reset eliminates all persisted message data in the store.
func (store *MemoryStore) Reset() {
	store.Lock()
	defer store.Unlock()
	if !store.opened {
		ERROR.Println(STR, "Trying to reset memory store, but not open")
		store.logger.Error("Trying to reset memory store, but not open", componentAttr(STR))
	}
	store.messages = make(map[string]packets.ControlPacket)
	WARN.Println(STR, "memorystore wiped")
	store.logger.Warn("memorystore wiped", componentAttr(STR))
}
