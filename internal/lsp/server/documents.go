package server

import (
	"maps"

	"github.com/backbone81/golr/internal/lsp/protocol"
)

// Documents holds the snapshots of the open documents at one moment. A change creates new Documents, so a value never
// changes.
type Documents struct {
	snapshots map[protocol.DocumentUri]*Snapshot
}

// Snapshot returns the snapshot of the document, or false when it is not open.
func (d *Documents) Snapshot(uri protocol.DocumentUri) (*Snapshot, bool) {
	snapshot, ok := d.snapshots[uri]
	return snapshot, ok
}

// with returns the documents with the snapshot added or replaced.
func (d *Documents) with(snapshot *Snapshot) *Documents {
	snapshots := maps.Clone(d.snapshots)
	snapshots[snapshot.URI] = snapshot
	return &Documents{snapshots: snapshots}
}

// without returns the documents without the document.
func (d *Documents) without(uri protocol.DocumentUri) *Documents {
	snapshots := maps.Clone(d.snapshots)
	delete(snapshots, uri)
	return &Documents{snapshots: snapshots}
}
