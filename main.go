package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math/cmplx"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
)

const (
	ViewWidth  = 640
	ViewHeight = 480
	MaxEscape  = 64
)

var palette []color.RGBA
var escapeColor color.RGBA

func init() {
	palette = make([]color.RGBA, MaxEscape)
	for i := 0; i < MaxEscape-1; i++ {
		palette[i] = color.RGBA{
			uint8(rand.Intn(256)),
			uint8(rand.Intn(256)),
			uint8(rand.Intn(256)),
			255}
	}
	escapeColor = color.RGBA{0, 0, 0, 0}
}

func escape(c complex128, a float64) int {
	z := c
	for i := 0; i < MaxEscape-1; i++ {
		if cmplx.Abs(z) > 2 {
			return i
		}
		z = complex(a, 0)*cmplx.Pow(z, 2) + c
	}
	return MaxEscape - 1
}

func generate(imgWidth int, imgHeight int, viewCenter complex128, radius float64, a float64) image.Image {
	m := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))
	zoomWidth := radius * 2
	pixelWidth := zoomWidth / float64(imgWidth)
	pixelHeight := pixelWidth
	viewHeight := (float64(imgHeight) / float64(imgWidth)) * zoomWidth
	left := (real(viewCenter) - (zoomWidth / 2)) + pixelWidth/2
	top := (imag(viewCenter) - (viewHeight / 2)) + pixelHeight/2

	var wgx sync.WaitGroup
	wgx.Add(imgWidth)
	for x := 0; x < imgWidth; x++ {
		go func(xx int) {
			defer wgx.Done()
			for y := 0; y < imgHeight; y++ {
				coord := complex(left+float64(xx)*pixelWidth, top+float64(y)*pixelHeight)
				f := escape(coord, a)
				if f == MaxEscape-1 {
					m.Set(xx, y, escapeColor)
				}
				m.Set(xx, y, palette[f])
			}
		}(x)
	}
	wgx.Wait()
	return m
}

func SafeFloat64(s string, def float64) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return def
	}
	return f
}

func pic(w http.ResponseWriter, r *http.Request) {
	mx := SafeFloat64(r.FormValue("mx"), 0.0)
	my := SafeFloat64(r.FormValue("my"), 0.0)
	radius := SafeFloat64(r.FormValue("radius"), 2.0)

	if r.Method == "GET" {
		http.ServeFile(w, r, "index.html") // <=== CONFLICTING?
	} else if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		// fmt.Fprintf(w, "Post from website! r.PostFrom = %v\n", r.PostForm)
		aValue := SafeFloat64(r.FormValue("a-value"), 1.0)
		// fmt.Fprintf(w, "Value = %s\n", aValue)
		m := generate(ViewWidth, ViewHeight, complex(mx, my), radius, aValue)
		w.Header().Set("Content-Type", "image/png")
		err := png.Encode(w, m)
		if err != nil {
			log.Println("png.Encode:", err)
		}
	} else {
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}

}

func main() {
	log.Println("Listening - open http://localhost:8000/ in browser")
	defer log.Println("Exiting")

	http.HandleFunc("/", pic)

	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatalf("ListenAndServe: %s\n", err)
	}
}
