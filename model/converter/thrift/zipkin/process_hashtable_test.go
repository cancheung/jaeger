// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package zipkin

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/uber/jaeger/model"
)

func TestProcessHashtable(t *testing.T) {
	p1 := model.NewProcess("s1", []model.KeyValue{
		model.String("host", "google.com"),
	})
	p1dup := model.NewProcess("s1", []model.KeyValue{
		model.String("host", "google.com"),
	})
	p2 := model.NewProcess("s2", []model.KeyValue{
		model.String("host", "facebook.com"),
	})
	ht := newProcessHashtable()
	assert.True(t, p1 == ht.add(p1))
	assert.True(t, p2 == ht.add(p2))
	assert.True(t, p1 == ht.add(p1))
	assert.True(t, p2 == ht.add(p2))
	assert.True(t, p1 == ht.add(p1dup))
}

func TestProcessHashtableCollision(t *testing.T) {
	ht := newProcessHashtable()
	// hash all processes to the same number
	ht.extHash = func(*model.Process) uint64 {
		return 42
	}

	p1 := model.NewProcess("s1", []model.KeyValue{
		model.String("host", "google.com"),
	})
	p2 := model.NewProcess("s2", []model.KeyValue{
		model.String("host", "facebook.com"),
	})
	assert.True(t, p1 == ht.add(p1))
	assert.True(t, p2 == ht.add(p2))
	assert.True(t, p1 == ht.add(p1))
	assert.True(t, p2 == ht.add(p2))
}
