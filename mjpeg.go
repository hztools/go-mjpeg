// {{{ Copyright (c) Paul R. Tagliamonte <paul@k3xec.com>, 2021
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
// THE SOFTWARE. }}}

package mjpeg

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"sync"
	"time"
)

// Stream contains a handle to the underlying image to be distributed to
// listening clients, as well as some internal state.
//
// This is thread-safe; any client may update or call methods on this object,
// and that will make its way out to all clients.
type Stream struct {
	opts Options
	lock *sync.RWMutex
	buf  *bytes.Buffer
}

// ServeHTTP will handle an HTTP request.
func (s *Stream) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.Body.Close() // We don't care about the user's data

	mw := multipart.NewWriter(w)
	defer mw.Close()

	// We don't use FormDataContentType since we want a multipart/x-mixed-replace.
	w.Header().Add("Content-Type", fmt.Sprintf(
		"multipart/x-mixed-replace;boundary=%s",
		mw.Boundary(),
	))
	w.WriteHeader(200)

	for {
		select {
		case <-time.After(s.opts.FrameDuration):
			w, err := mw.CreatePart(textproto.MIMEHeader{
				"Content-Type": []string{"image/jpeg"},
			})
			if err != nil {
				// Log this
				return
			}

			s.lock.RLock()
			w.Write(s.buf.Bytes())
			s.lock.RUnlock()
		case <-s.opts.Context.Done(): // Parent context is done.
			return
		case <-r.Context().Done(): // Request context is done.
			return
		}
	}
}

// Update will set the current frame to the provided Image. This will preform
// a JPEG encoding, and stream those bytes -- this image handle will not be
// held by the Stream after this call.
func (s *Stream) Update(i image.Image) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.buf.Reset()
	if err := jpeg.Encode(s.buf, i, nil); err != nil {
		return err
	}

	// TODO(paultag): Copy s.buf.Bytes out to a buffer?

	return nil
}

// NewStream will return a new Stream object, with default options. If
// control over timings, etc is required, you may use the NewStreamWithOptions
// helper.
func NewStream() *Stream {
	return NewStreamWithOptions(Options{
		FrameDuration: time.Second / 20,
		Context:       context.Background(),
	})
}

// Options contains all the knobs exposed from the library.
type Options struct {
	// FrameDuration determines how many frames per second are sent to the
	// client.
	FrameDuration time.Duration

	// Context is the root context -- not related to the context of the
	// per-connection HTTP streams, which is pulled from the http.Request
	Context context.Context
}

// NewStreamWithOptions will return a new Stream object, with the options
// specified by the caller.
func NewStreamWithOptions(opts Options) *Stream {
	return &Stream{
		opts: opts,
		lock: &sync.RWMutex{},
		buf:  &bytes.Buffer{},
	}
}

// vim: foldmethod=marker
