// Copyright 2023-2024 The NATS Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package stree

// Node with 4 children
type node4 struct {
	meta
	children [4]node
	keys     [4]byte
}

func newNode4(prefix []byte) *node4 {
	nn := &node4{}
	nn.setPrefix(prefix)
	return nn
}

func (n *node4) isLeaf() bool { return false }
func (n *node4) base() *meta  { return &n.meta }

func (n *node4) setPrefix(pre []byte) {
	n.prefixLen = uint16(min(len(pre), maxPrefixLen))
	for i := uint16(0); i < n.prefixLen; i++ {
		n.prefix[i] = pre[i]
	}
}

// Currently we do not need to keep sorted for traversal so just add to the end.
func (n *node4) addChild(c byte, nn node) {
	if n.size >= 4 {
		panic("node4 full!")
	}
	n.keys[n.size] = c
	n.children[n.size] = nn
	n.size++
}

func (n *node4) numChildren() uint16 { return n.size }
func (n *node4) path() []byte        { return n.prefix[:n.prefixLen] }

func (n *node4) findChild(c byte) *node {
	for i := uint16(0); i < n.size; i++ {
		if n.keys[i] == c {
			return &n.children[i]
		}
	}
	return nil
}

func (n *node4) isFull() bool { return n.size >= 4 }

func (n *node4) grow() node {
	nn := newNode16(n.prefix[:n.prefixLen])
	for i := 0; i < 4; i++ {
		nn.addChild(n.keys[i], n.children[i])
	}
	return nn
}

// Deletes a child from the node.
func (n *node4) deleteChild(c byte) {
	for i, last := uint16(0), n.size-1; i < n.size; i++ {
		if n.keys[i] == c {
			// Unsorted so just swap in last one here, else nil if last.
			if i < last {
				n.keys[i] = n.keys[last]
				n.children[i] = n.children[last]
				n.keys[last] = 0
				n.children[last] = nil
			} else {
				n.keys[i] = 0
				n.children[i] = nil
			}
			n.size--
			return
		}
	}
}

// Shrink if needed and return new node, otherwise return nil.
func (n *node4) shrink() node {
	if n.size == 1 {
		return n.children[0]
	}
	return nil
}

// Will match parts against our prefix.
func (n *node4) matchParts(parts [][]byte) ([][]byte, bool) {
	return matchParts(parts, n.prefix[:n.prefixLen])
}

// Iterate over all children calling func f.
func (n *node4) iter(f func(node) bool) {
	for i := uint16(0); i < n.size; i++ {
		if !f(n.children[i]) {
			return
		}
	}
}
