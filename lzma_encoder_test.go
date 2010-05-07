// Copyright (c) 2010, Andrei Vieru. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lzma

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"reflect"
	"testing"
)

func pipe(t *testing.T, efunc func(io.WriteCloser), dfunc func(io.ReadCloser), size int64) {
	level := 3
	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		ze := NewEncoderSizeLevel(pw, size, level)
		defer ze.Close()
		efunc(ze)
	}()
	defer pr.Close()
	zd := NewDecoder(pr)
	defer zd.Close()
	dfunc(zd)
}

func testEmpty(t *testing.T, sizeIsKnown bool) {
	size := int64(-1)
	if sizeIsKnown == true {
		size = 0
	}
	pipe(t,
		func(w io.WriteCloser) {},
		func(r io.ReadCloser) {
			b, err := ioutil.ReadAll(r)
			if err != nil {
				t.Fatalf("%v", err)
			}
			if len(b) != 0 {
				t.Fatalf("did not read an empty slice")
			}
		},
		size)
}

func TestEmpty1(t *testing.T) {
	testEmpty(t, true)
}

func TestEmpty2(t *testing.T) {
	testEmpty(t, false)
}

func testBoth(t *testing.T, sizeIsKnown bool) {
	size := int64(-1)
	payload := []byte("lzmalzmalzma")
	if sizeIsKnown == true {
		size = int64(len(payload))
	}
	pipe(t,
		func(w io.WriteCloser) {
			n, err := w.Write(payload)
			if err != nil {
				t.Fatalf("%v", err)
			}
			if n != len(payload) {
				t.Fatalf("wrote %d bytes, want %d bytes", n, len(payload))
			}
		},
		func(r io.ReadCloser) {
			b, err := ioutil.ReadAll(r)
			if err != nil {
				t.Fatalf("%v", err)
			}
			if string(b) != string(payload) {
				t.Fatalf("payload is %s, want %s", string(b), string(payload))
			}
		},
		size)
}

func TestBoth1(t *testing.T) {
	testBoth(t, true)
}

func TestBoth2(t *testing.T) {
	testBoth(t, false)
}

func TestEncoder(t *testing.T) {
	b := new(bytes.Buffer)
	for _, tt := range unlzmaTests {
		if tt.err == nil {
			pr, pw := io.Pipe()
			defer pr.Close()
			in := bytes.NewBuffer([]byte(tt.raw))
			size := int64(-1)
			if tt.size == true {
				size = int64(len([]byte(tt.raw)))
			}
			go func() {
				defer pw.Close()
				w := NewEncoderSizeLevel(pw, size, tt.level)
				defer w.Close()
				_, err := io.Copy(w, in)
				if err != nil {
					t.Errorf("%v", err)
				}
			}()
			b.Reset()
			_, err := io.Copy(b, pr)
			if err != nil {
				t.Errorf("%v", err)
			}
			res := b.Bytes()
			if reflect.DeepEqual(res, tt.lzma) == false {
				t.Errorf("%s: got %d-byte %q, want %d-byte %q", tt.descr, len(res), string(res), len(tt.lzma), string(tt.lzma))
			}
		}
	}
}

func BenchmarkEncoder(b *testing.B) {
	buf := new(bytes.Buffer)
	for i := 0; i < b.N; i++ {
		in := bytes.NewBuffer([]byte(bk.raw))
		pr, pw := io.Pipe()
		go func() {
			w := NewEncoderLevel(pw, bk.level)
			_, err := io.Copy(w, in)
			if err != nil {
				log.Exitf("%v", err)
			}
			defer pw.Close()
			defer w.Close()
		}()
		buf.Reset()
		_, err := io.Copy(buf, pr)
		if err != nil {
			log.Exitf("%v", err)
		}
		defer pr.Close()
	}
	if reflect.DeepEqual(buf.Bytes(), bk.lzma) == false { // check only once, not at every iteration
		log.Exitf("%s: got %d-byte %q, want %d-byte %q", bk.descr, len(buf.Bytes()), buf.String(), len(bk.lzma), string(bk.lzma))
	}
}
