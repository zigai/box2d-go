// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import "github.com/zigai/box2d-go/internal/core"

const (
	nullIndex int = -1
	// Simple djb2 hash function for determinism testing.
	HashInit uint32 = 5381
	// Significant API and behavior changes.
	VersionMajor int = 3
	// API changes and features.
	VersionMinor int = 2
	// Bug fixes. No API changes.
	VersionRevision int = 0
)

// Version numbering scheme.
type Version struct {
	Major    int
	Minor    int
	Revision int
}

// AssertError is the panic value for invalid API input and, unless built with the
// box2d_noassert tag, for broken internal invariants.
type AssertError = core.AssertError

// Get the current version of Box2D.
func CurrentVersion() Version {
	return Version{Major: VersionMajor, Minor: VersionMinor, Revision: VersionRevision}
}

// Return true if the library was built with box2d_double (large world mode).
func IsDoublePrecision() bool {
	return doublePrecision
}

// Simple djb2 hash function for determinism testing.
func Hash(hash uint32, data []byte) uint32 {
	for _, b := range data {
		hash = ((hash << 5) + hash) + uint32(b)
	}

	return hash
}
