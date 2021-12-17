package main

import (
	"fmt"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
	"github.com/gorilla/mux"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"net/http"
	"os"
	"runtime"
	"strconv"
)

func generateLineItems() []opts.LineData {
	var x = x0
	items := make([]opts.LineData, 0)
	for i := 0; i < xMax; i++ {
		if i > 0 {
			x = r * x * (1 - x)
		}
		// log.Println(x)
		items = append(items, opts.LineData{Value: x})
	}
	return items
}

func httpserver(w http.ResponseWriter, req *http.Request) {

	key := req.URL.Query().Get("x0")
	if s, err := strconv.ParseFloat(key, 64); err == nil {
		x0 = s
	} else {
		x0 = 0.4
	}

	key = req.URL.Query().Get("r")
	if s, err := strconv.ParseFloat(key, 64); err == nil {
		r = s
	} else {
		r = 2.6
	}

	key = req.URL.Query().Get("years")
	if s, err := strconv.Atoi(key); err == nil {
		xMax = s
	} else {
		xMax = 30
	}

	var xArr []string
	for i := 1; i <= xMax; i++ {
		xArr = append(xArr, fmt.Sprintf("%v", i))
	}

	line := charts.NewLine()

	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeMacarons}),
		charts.WithTitleOpts(opts.Title{
			Title:    "График популяции",
			Subtitle: fmt.Sprintf("Первичная популяция: x0 = %v; Постоянная роста: r = %v; Года: years = %v; (параметры можно задать следующим образом: '?years=30&x0=0.4&r=2.6')", x0, r, xMax),
		}))
	line.SetXAxis(xArr).
		AddSeries("Category A", generateLineItems()).
		SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: true}))
	err := line.Render(w)
	if err != nil {
		return
	}
}

var x0 float64 //Алғашқы популяция
var r float64  //Өсу қарқыны
var xMax int

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/", httpserver)
	router.HandleFunc("/diagram", httpserver2)
	http.Handle("/", router)
	port := os.Getenv("PORT")
	if port == "" {
		fmt.Println("Port env not found, set 8081")
		port = "8081"
	} else {
		fmt.Println("Port env = " + port)
	}
	fmt.Println("Running on http://127.0.0.1:" + port + "/?years=30&x0=0.4&r=2.6 and http://127.0.0.1:" + port + "/diagram?w=700&h=300")
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		return
	}
}

//
func httpserver2(w http.ResponseWriter, req *http.Request) {

	var canvasHeight int
	var canvasWidth int

	key := req.URL.Query().Get("h")
	if s, err := strconv.Atoi(key); err == nil {
		canvasHeight = s
	} else {
		canvasHeight = 1500
	}

	key = req.URL.Query().Get("w")
	if s, err := strconv.Atoi(key); err == nil {
		canvasWidth = s
	} else {
		canvasWidth = 3500
	}

	//
	upLeft := image.Point{}
	lowRight := image.Point{X: canvasWidth, Y: canvasHeight}
	gr := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	im := image.NewRGBA(image.Rectangle{Min: upLeft, Max: lowRight})
	draw.Draw(im, im.Bounds(), &image.Uniform{C: gr}, image.Point{}, draw.Src)

	var numCPU = runtime.NumCPU()
	stepSize := (1.0 - 0.01) / float64(canvasWidth)
	sliceSize := 1.0 / (float64)(numCPU)
	c := make(chan bool, numCPU)

	for cpu := 0; cpu < numCPU; cpu++ {
		from := (float64)(cpu) * sliceSize
		to := (float64)(cpu+1) * sliceSize
		go computeX(im, stepSize, from, to, c)
	}
	for cpu := 0; cpu < numCPU; cpu++ {
		<-c // wait
	}

	w.Header().Set("Content-Type", "image/png")
	err := png.Encode(w, im)
	if err != nil {
		return
	}
}

func logistic(x float64, r float64, it int) float64 {
	for i := 0; i < it; i++ {
		x = r * x * (1 - x)
	}
	return x
}

func logr(im *image.RGBA, x float64, its int) {
	var i, j = 0, 0

	var X, Y = (float64)(im.Bounds().Max.X), (float64)(im.Bounds().Max.Y)
	steps := 4.0 / X
	for r := 0.0; r < 4.0; r += steps {
		res := logistic(x, r, its)
		i = (int)(X * (r - 0.0) / 4.0)
		j = (int)(Y - Y*res)
		k := im.RGBAAt(i, j)
		if k.A == 255 {
			if k.R > 1 {
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
	for its := 50; its < 150; its += 10 {
		for x := from; x < to; x += stepSize {
			go logr(im, x, its)
		}
	}
	c <- true
}
