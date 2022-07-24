// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file is taken from crypto/internal/randutil in the go1.14.4 standard
// library.

package passit

import (
	"io"
	"sync"
)

var (
	closedChanOnce sync.Once
	closedChan     chan struct{}
)

// maybeReadByte reads a single byte from r with ~50% probability. This is used
// to ensure that callers do not depend on non-guaranteed behaviour, e.g.
// assuming that rsa.GenerateKey is deterministic w.r.t. a given random stream.
//
// This does not affect tests that pass a stream of fixed bytes as the random
// source (e.g. a zeroReader).
func maybeReadByte(r io.Reader) {
	closedChanOnce.Do(func() {
		closedChan = make(chan struct{})
		close(closedChan)
	})

	select {
	case <-closedChan:
		return
	case <-closedChan:
		var buf [1]byte
		r.Read(buf[:])
	}
}
