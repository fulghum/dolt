// Copyright 2021 Dolthub, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package prolly

import (
	"context"

	"github.com/dolthub/dolt/go/store/hash"
	"github.com/dolthub/dolt/go/store/pool"
	"github.com/dolthub/dolt/go/store/val"
)

func fetchChild(ctx context.Context, ns NodeStore, mt metaValue) (Node, error) {
	// todo(andy) handle nil Node, dangling ref
	return ns.Read(ctx, mt.GetRef())
}

func writeNewChild(ctx context.Context, ns NodeStore, level uint64, items ...nodeItem) (Node, nodePair, error) {
	child := makeProllyNode(ns.Pool(), level, items...)

	ref, err := ns.Write(ctx, child)
	if err != nil {
		return nil, nodePair{}, err
	}

	if len(items) == 0 {
		// empty leaf node
		return child, nodePair{}, nil
	}

	lastKey := val.Tuple(items[len(items)-metaPairCount])
	metaKey := val.CloneTuple(ns.Pool(), lastKey)
	metaVal := newMetaValue(ns.Pool(), child.cumulativeCount(), ref)
	meta := nodePair{nodeItem(metaKey), nodeItem(metaVal)}

	return child, meta, nil
}

const (
	metaPairCount  = 2
	metaPairKeyIdx = 0
	metaPairValIdx = 1

	metaValueCountIdx = 0
	metaValueRefIdx   = 1
)

// metaValue is a value Tuple in an internal Node of a prolly tree.
// metaValues have two fields: cumulative count and ref.
type metaValue val.Tuple

func newMetaValue(pool pool.BuffPool, count uint64, ref hash.Hash) metaValue {
	var cnt [6]byte
	val.WriteUint48(cnt[:], count)
	return metaValue(val.NewTuple(pool, cnt[:], ref[:]))
}

// GetCumulativeCount returns the cumulative number of nodeItems
// within the subtree pointed to by a metaValue.
func (mt metaValue) GetCumulativeCount() uint64 {
	cnt := val.Tuple(mt).GetField(metaValueCountIdx)
	return val.ReadUint48(cnt)
}

// GetRef returns the hash.Hash of the child Node pointed
// to by this metaValue.
func (mt metaValue) GetRef() hash.Hash {
	tup := val.Tuple(mt)
	ref := tup.GetField(metaValueRefIdx)
	return hash.New(ref)
}
