// Package vdf bridges github.com/gofurry/vdf-go into the steam-go addon tree.
//
// It is a thin compatibility layer for callers who want to work with Valve Data
// Format (VDF / KeyValues) text files while already depending on steam-go.
//
// This package does not re-implement the parser. It re-exports the stable
// vdf-go API and keeps steam-go focused on Steam Web API and addon integration.
//
// Scope:
//
//   - text VDF / KeyValues files
//   - appmanifest_*.acf
//   - libraryfolders.vdf
//   - config.vdf / loginusers.vdf style text files
//
// Non-goals:
//
//   - binary VDF
//   - shortcuts.vdf
//   - automatic Steam installation scanning
//   - reading user directories automatically
//   - account/session extraction
package vdf
