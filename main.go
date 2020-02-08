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

type Fractal struct {
	a, b, d, e float64
}

func escape(c complex128, fractal Fractal) int {
	z := c
	for i := 0; i < MaxEscape-1; i++ {
		if cmplx.Abs(z) > 2 {
			return i
		}
		z = complex(fractal.a, fractal.b)*cmplx.Pow(z, complex(fractal.d, fractal.e)) + c
	}
	return MaxEscape - 1
}

func generate(imgWidth int, imgHeight int, viewCenter complex128, radius float64, fractal Fractal) image.Image {
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
				f := escape(coord, fractal)
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

func index(w http.ResponseWriter, r *http.Request) {
	mx := SafeFloat64(r.FormValue("mx"), 0.0)
	my := SafeFloat64(r.FormValue("my"), 0.0)
	radius := SafeFloat64(r.FormValue("radius"), 2.0)

	pic := func(fractal Fractal) {
		m := generate(ViewWidth, ViewHeight, complex(mx, my), radius, fractal)
		w.Header().Set("Content-Type", "image/png")
		err := png.Encode(w, m)
		if err != nil {
			log.Println("png.Encode:", err)
		}
	}

	switch r.Method {
	case "GET":
		http.ServeFile(w, r, "index.html") // <=== CONFLICTING?
		startFractal := Fractal{1.0, 2.0, 0.0, 0.0}
		pic(startFractal)
	case "POST":
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		userFractal := Fractal{
			SafeFloat64(r.FormValue("a-value"), 1.0),
			SafeFloat64(r.FormValue("b-value"), 2.0),
			SafeFloat64(r.FormValue("d-value"), 0.0),
			SafeFloat64(r.FormValue("e-value"), 0.0),
		}
		pic(userFractal)
	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}
}

func main() {
	log.Println("Listening - open http://localhost:8000/ in browser")
	defer log.Println("Exiting")

	http.HandleFunc("/", index)

	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatalf("ListenAndServe: %s\n", err)
	}
}
