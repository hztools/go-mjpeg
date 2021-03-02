package main

import (
	"image"
	"log"
	"math/rand"
	"net/http"
	"time"

	"hz.tools/mjpeg"
)

func main() {
	i := image.NewGray(image.Rectangle{
		Min: image.Point{X: 0, Y: 0},
		Max: image.Point{X: 500, Y: 500},
	})

	stream := mjpeg.NewStream()
	go func() {
		for {

			for j := range i.Pix {
				i.Pix[j] = uint8(rand.Uint32())
			}

			time.Sleep(time.Second / 30)
			stream.Update(i)
		}
	}()

	http.Handle("/", stream)
	log.Printf("Listening for requests to http://localhost:8888")
	http.ListenAndServe("127.0.0.1:8888", nil)
}
