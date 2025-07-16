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
	"io/fs"
	"log/slog"
	"os"
	"path"
	"sort"
	"sync"

	"github.com/eclipse/paho.mqtt.golang/packets"
)

const (
	msgExt     = ".msg"
	tmpExt     = ".tmp"
	corruptExt = ".CORRUPT"
)

// FileStore implements the store interface using the filesystem to provide
// true persistence, even across client failure. This is designed to use a
// single directory per running client. If you are running multiple clients
// on the same filesystem, you will need to be careful to specify unique
// store directories for each.
type FileStore struct {
	sync.RWMutex
	directory string
	opened    bool
	logger    *slog.Logger
}

// NewFileStore will create a new FileStore which stores its messages in the
// directory provided.
func NewFileStore(directory string) *FileStore {
	store := &FileStore{
		directory: directory,
		opened:    false,
		logger:    noopSLogger,
	}
	return store
}

// NewFileStoreEx will create a new FileStore which stores its messages in the
// directory provided, using the provided logger.
func NewFileStoreEx(directory string, logger *slog.Logger) *FileStore {
	if logger == nil {
		logger = noopSLogger
	}
	store := &FileStore{
		directory: directory,
		opened:    false,
		logger:    logger,
	}
	return store
}

// Open will allow the FileStore to be used.
func (store *FileStore) Open() {
	store.Lock()
	defer store.Unlock()
	// if no store directory was specified in ClientOpts, by default use the
	// current working directory
	if store.directory == "" {
		store.directory, _ = os.Getwd()
	}

	// if store dir exists, great, otherwise, create it
	if !exists(store.directory) {
		perms := os.FileMode(0770)
		merr := os.MkdirAll(store.directory, perms)
		chkerr(merr)
	}
	store.opened = true
	DEBUG.Println(STR, "store is opened at", store.directory)
	store.logger.Debug("store is opened", slog.String("directory", store.directory), componentAttr(STR))
}

// Close will disallow the FileStore from being used.
func (store *FileStore) Close() {
	store.Lock()
	defer store.Unlock()
	store.opened = false
	DEBUG.Println(STR, "store is closed")
	store.logger.Debug("store is closed", componentAttr(STR))
}

// Put will put a message into the store, associated with the provided
// key value.
func (store *FileStore) Put(key string, m packets.ControlPacket) {
	store.Lock()
	defer store.Unlock()
	if !store.opened {
		ERROR.Println(STR, "Trying to use file store, but not open")
		store.logger.Error("Trying to use file store, but not open", componentAttr(STR))
		return
	}
	full := fullpath(store.directory, key)
	write(store.directory, key, m)
	if !exists(full) {
		ERROR.Println(STR, "file not created:", full)
		store.logger.Error("file not created", slog.String("path", full), componentAttr(STR))
	}
}

// Get will retrieve a message from the store, the one associated with
// the provided key value.
func (store *FileStore) Get(key string) packets.ControlPacket {
	store.RLock()
	defer store.RUnlock()
	if !store.opened {
		ERROR.Println(STR, "trying to use file store, but not open")
		store.logger.Error("trying to use file store, but not open", componentAttr(STR))
		return nil
	}
	filepath := fullpath(store.directory, key)
	if !exists(filepath) {
		return nil
	}
	mfile, oerr := os.Open(filepath)
	chkerr(oerr)
	msg, rerr := packets.ReadPacket(mfile)
	chkerr(mfile.Close())

	// Message was unreadable, return nil
	if rerr != nil {
		newpath := corruptpath(store.directory, key)
		WARN.Println(STR, "corrupted file detected:", rerr.Error(), "archived at:", newpath)
		store.logger.Warn("corrupted file detected", slog.String("error", rerr.Error()), slog.String("archived at", newpath), componentAttr(STR))
		if err := os.Rename(filepath, newpath); err != nil {
			ERROR.Println(STR, err)
			store.logger.Error("failed to archive corrupted file", slog.String("error", err.Error()), componentAttr(STR))
		}
		return nil
	}
	return msg
}

// All will provide a list of all of the keys associated with messages
// currently residing in the FileStore.
func (store *FileStore) All() []string {
	store.RLock()
	defer store.RUnlock()
	return store.all()
}

// Del will remove the persisted message associated with the provided
// key from the FileStore.
func (store *FileStore) Del(key string) {
	store.Lock()
	defer store.Unlock()
	store.del(key)
}

// Reset will remove all persisted messages from the FileStore.
func (store *FileStore) Reset() {
	store.Lock()
	defer store.Unlock()
	WARN.Println(STR, "FileStore Reset")
	store.logger.Warn("FileStore Reset", componentAttr(STR))
	for _, key := range store.all() {
		store.del(key)
	}
}

// lockless
func (store *FileStore) all() []string {
	var err error
	var keys []string

	if !store.opened {
		ERROR.Println(STR, "trying to use file store, but not open")
		store.logger.Error("trying to use file store, but not open", componentAttr(STR))
		return nil
	}

	entries, err := os.ReadDir(store.directory)
	chkerr(err)
	files := make(fileInfos, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		chkerr(err)
		files = append(files, info)
	}
	sort.Sort(files)
	for _, f := range files {
		DEBUG.Println(STR, "file in All():", f.Name())
		store.logger.Debug("file in All()", slog.String("name", f.Name()), componentAttr(STR))
		name := f.Name()
		if len(name) < len(msgExt) || name[len(name)-len(msgExt):] != msgExt {
			DEBUG.Println(STR, "skipping file, doesn't have right extension: ", name)
			store.logger.Debug("skipping file, doesn't have right extension", slog.String("name", name), componentAttr(STR))
			continue
		}
		key := name[0 : len(name)-4] // remove file extension
		keys = append(keys, key)
	}
	return keys
}

// lockless
func (store *FileStore) del(key string) {
	if !store.opened {
		ERROR.Println(STR, "trying to use file store, but not open")
		store.logger.Error("trying to use file store, but not open", componentAttr(STR))
		return
	}
	DEBUG.Println(STR, "store del filepath:", store.directory)
	store.logger.Debug("store del filepath", slog.String("directory", store.directory), componentAttr(STR))
	DEBUG.Println(STR, "store delete key:", key)
	store.logger.Debug("store delete key", slog.String("key", key), componentAttr(STR))
	filepath := fullpath(store.directory, key)
	DEBUG.Println(STR, "path of deletion:", filepath)
	store.logger.Debug("path of deletion", slog.String("filepath", filepath), componentAttr(STR))
	if !exists(filepath) {
		WARN.Println(STR, "store could not delete key:", key)
		store.logger.Warn("store could not delete key", slog.String("key", key), componentAttr(STR))
		return
	}
	rerr := os.Remove(filepath)
	chkerr(rerr)
	DEBUG.Println(STR, "del msg:", key)
	store.logger.Debug("del msg", slog.String("key", key), componentAttr(STR))
	if exists(filepath) {
		ERROR.Println(STR, "file not deleted:", filepath)
		store.logger.Error("file not deleted", slog.String("filepath", filepath), componentAttr(STR))
	}
}

func fullpath(store string, key string) string {
	p := path.Join(store, key+msgExt)
	return p
}

func tmppath(store string, key string) string {
	p := path.Join(store, key+tmpExt)
	return p
}

func corruptpath(store string, key string) string {
	p := path.Join(store, key+corruptExt)
	return p
}

// create file called "X.[messageid].tmp" located in the store
// the contents of the file is the bytes of the message, then
// rename it to "X.[messageid].msg", overwriting any existing
// message with the same id
// X will be 'i' for inbound messages, and O for outbound messages
func write(store, key string, m packets.ControlPacket) {
	temppath := tmppath(store, key)
	f, err := os.Create(temppath)
	chkerr(err)
	werr := m.Write(f)
	chkerr(werr)
	cerr := f.Close()
	chkerr(cerr)
	rerr := os.Rename(temppath, fullpath(store, key))
	chkerr(rerr)
}

func exists(file string) bool {
	if _, err := os.Stat(file); err != nil {
		if os.IsNotExist(err) {
			return false
		}
		chkerr(err)
	}
	return true
}

type fileInfos []fs.FileInfo

func (f fileInfos) Len() int {
	return len(f)
}

func (f fileInfos) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func (f fileInfos) Less(i, j int) bool {
	return f[i].ModTime().Before(f[j].ModTime())
}
