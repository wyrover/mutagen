// Subset of https://github.com/rjeczalik/notify extracted and modified to
// expose watcher functionality directly. Originally extracted from the
// following revision:
// https://github.com/rjeczalik/notify/tree/52ae50d8490436622a8941bd70c3dbe0acdd4bbf
//
// The original code license:
//
// The MIT License (MIT)
//
// Copyright (c) 2014-2015 The Notify Authors
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
//
// The original license header inside the code itself:
//
// Copyright (c) 2014-2015 The Notify Authors. All rights reserved.
// Use of this source code is governed by the MIT license that can be
// found in the LICENSE file.

package notify

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const all = ^Event(0)
const sep = string(os.PathSeparator)

var errDepth = errors.New("exceeded allowed iteration count (circular symlink?)")

func min(i, j int) int {
	if i > j {
		return j
	}
	return i
}

func max(i, j int) int {
	if i < j {
		return j
	}
	return i
}

// must panics if err is non-nil.
func must(err error) {
	if err != nil {
		panic(err)
	}
}

// nonil gives first non-nil error from the given arguments.
func nonil(err ...error) error {
	for _, err := range err {
		if err != nil {
			return err
		}
	}
	return nil
}

func cleanpath(path string) (realpath string, isrec bool, err error) {
	if strings.HasSuffix(path, "...") {
		isrec = true
		path = path[:len(path)-3]
	}
	if path, err = filepath.Abs(path); err != nil {
		return "", false, err
	}
	if path, err = canonical(path); err != nil {
		return "", false, err
	}
	return path, isrec, nil
}

// canonical resolves any symlink in the given path and returns it in a clean form.
// It expects the path to be absolute. It fails to resolve circular symlinks by
// maintaining a simple iteration limit.
func canonical(p string) (string, error) {
	p, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	for i, j, depth := 1, 0, 1; i < len(p); i, depth = i+1, depth+1 {
		if depth > 128 {
			return "", &os.PathError{Op: "canonical", Path: p, Err: errDepth}
		}
		if j = strings.IndexRune(p[i:], '/'); j == -1 {
			j, i = i, len(p)
		} else {
			j, i = i, i+j
		}
		fi, err := os.Lstat(p[:i])
		if err != nil {
			return "", err
		}
		if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
			s, err := os.Readlink(p[:i])
			if err != nil {
				return "", err
			}
			if filepath.IsAbs(s) {
				p = "/" + s + p[i:]
			} else {
				p = p[:j] + s + p[i:]
			}
			i = 1 // no guarantee s is canonical, start all over
		}
	}
	return filepath.Clean(p), nil
}

func joinevents(events []Event) (e Event) {
	if len(events) == 0 {
		e = All
	} else {
		for _, event := range events {
			e |= event
		}
	}
	return
}

func split(s string) (string, string) {
	if i := lastIndexSep(s); i != -1 {
		return s[:i], s[i+1:]
	}
	return "", s
}

func base(s string) string {
	if i := lastIndexSep(s); i != -1 {
		return s[i+1:]
	}
	return s
}

func indexbase(root, name string) int {
	if n, m := len(root), len(name); m >= n && name[:n] == root &&
		(n == m || name[n] == os.PathSeparator) {
		return min(n+1, m)
	}
	return -1
}

func indexSep(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == os.PathSeparator {
			return i
		}
	}
	return -1
}

func lastIndexSep(s string) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == os.PathSeparator {
			return i
		}
	}
	return -1
}
