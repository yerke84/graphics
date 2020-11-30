package main

import (
    "net/http"
    "fmt"
    "strconv"
    "github.com/go-echarts/go-echarts/v2/charts"
    "github.com/go-echarts/go-echarts/v2/opts"
    "github.com/go-echarts/go-echarts/v2/types"
)

func generateLineItems() []opts.LineData {
    var x float64 = x0;
    items := make([]opts.LineData, 0)
    for i := 0; i < xMax; i++ {
      if(i > 0) {
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
    }

    key = req.URL.Query().Get("r")
    if s, err := strconv.ParseFloat(key, 64); err == nil {
        r = s
    }

    key = req.URL.Query().Get("years")
    if s, err := strconv.Atoi(key); err == nil {
        xMax = s
    }

    var xArr []string;
    for i := 1; i <= xMax; i++ {
      xArr = append(xArr, fmt.Sprintf("%v", i))
    }

    line := charts.NewLine()

    line.SetGlobalOptions(
        charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeMacarons}),
       charts.WithTitleOpts(opts.Title{
            Title:    "Популяция графигі",
           Subtitle: fmt.Sprintf("Алғашқы популяция: x0 = %v; Өсу қарқыны: r = %v; Жылдар: years = %v", x0, r, xMax),
       }))
    line.SetXAxis(xArr).
      AddSeries("Category A", generateLineItems()).
        SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: true}))
    line.Render(w)
}

var x0 float64; //Алғашқы популяция
var r float64; //Өсу қарқыны
var xMax int;

func main() {
    fmt.Println("Running on http://127.0.0.1:8081/?years=30&x0=0.4&r=2.6")
    http.HandleFunc("/", httpserver)
    http.ListenAndServe(":8081", nil)
}
