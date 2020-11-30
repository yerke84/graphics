package main

import (
	"image"
  "net/http"
  "fmt"
	"image/png"
	"image/color"
	"image/draw"
	"runtime"
  "strconv"
)

func logistic(x float64, r float64, it int) float64 {
	for i := 0; i < it; i++ {
		x = r * x * (1 - x)
	}
	return x
}

func logr(im *image.RGBA, x float64, its int) {
	var i,j int = 0,0

	var X,Y float64= (float64)(im.Bounds().Max.X), (float64)(im.Bounds().Max.Y)
	steps := 4.0 / X
	for r := 0.0; r < 4.0 ; r += steps {
		res := logistic(x, r, its)
		i = (int)((X * (r-0.0)/4.0) )
		j = (int)((Y - Y * res))
		k := im.RGBAAt(i,j)
		if (k.A == 255) {
			if (k.R > 1) {
				k.R = k.R - 1
				k.G = k.G - 1
				k.B = k.B - 1
			}
		} else {
			k.R = 175
			k.G = 175
			k.B = 175
			k.A = 255
		}
		im.Set(i, j, k)
	}
}

func computeX(im *image.RGBA, stepSize float64, from float64, to float64, c chan bool) {
	for its := 50 ; its < 150 ; its += 10 {
		for x := from; x < to; x += stepSize {
			logr(im, x, its)
		}
	}
	c <- true
}

func httpserver(w http.ResponseWriter, req *http.Request) {

    var canvasHeight int
    var canvasWidth int

    key := req.URL.Query().Get("h")
    if s, err := strconv.Atoi(key); err == nil {
        canvasHeight = s
    }

    key = req.URL.Query().Get("w")
    if s, err := strconv.Atoi(key); err == nil {
        canvasWidth = s
    }

    //
    upLeft := image.Point{0, 0}
    lowRight := image.Point{canvasWidth, canvasHeight}
		gr := color.RGBA{255,255,255,255}
		im := image.NewRGBA(image.Rectangle{upLeft, lowRight})
		draw.Draw(im, im.Bounds(), &image.Uniform{gr}, image.ZP, draw.Src)

		var numCPU = runtime.NumCPU()
		stepSize := (1.0-0.01) / float64(canvasWidth)
		sliceSize := 1.0/(float64)(numCPU)
		c := make(chan bool, numCPU)

		for cpu := 0 ; cpu < numCPU; cpu++ {
			from := (float64)(cpu) * sliceSize
			to := (float64)(cpu+1) * sliceSize
			go computeX(im, stepSize, from, to , c)
		}
		for cpu := 0; cpu < numCPU; cpu++ {
			<-c    // wait
		}

		w.Header().Set("Content-Type", "image/png")
    png.Encode(w, im)
}

func main() {
  fmt.Println("Running on http://127.0.0.1:8083/?w=7000&h=3000")
  http.HandleFunc("/", httpserver)
  http.ListenAndServe(":8083", nil)
}
