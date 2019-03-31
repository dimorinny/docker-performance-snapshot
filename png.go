package main

import (
	"bufio"
	"fmt"
	"github.com/wcharczuk/go-chart"
)

type PngReporter struct {
	directoryPath string
}

func NewPngReporter(directoryPath string) Reporter {
	return &PngReporter{
		directoryPath: directoryPath,
	}
}

func (r *PngReporter) Report(name string, description string, metrics []*Metric) error {
	pngFile, err := createFile(
		r.directoryPath,
		fmt.Sprintf("%s.png", name),
	)
	if err != nil {
		return err
	}
	defer func() { _ = pngFile.Close() }()

	pngMetricsWriter := bufio.NewWriter(pngFile)

	var (
		xValues []float64
		yValues []float64
	)

	for _, metric := range metrics {
		xValues = append(xValues, float64(metric.Timestamp))
		yValues = append(yValues, metric.Value)
	}

	graph := chart.Chart{
		Title:      description,
		TitleStyle: chart.StyleShow(),
		Height:     1400,
		Background: chart.Style{
			Padding: chart.Box{
				Top:    145,
				Bottom: 15,
				Left:   15,
				Right:  15,
			},
		},
		XAxis: chart.XAxis{
			Style:     chart.StyleShow(),
			Name:      "time",
			NameStyle: chart.StyleShow(),
		},
		YAxis: chart.YAxis{
			Style:     chart.StyleShow(),
			Name:      name,
			NameStyle: chart.StyleShow(),
		},
		Series: []chart.Series{
			chart.ContinuousSeries{
				Style: chart.Style{
					Show:        true,
					StrokeColor: chart.GetDefaultColor(0).WithAlpha(64),
					FillColor:   chart.GetDefaultColor(0).WithAlpha(64),
				},
				XValues: xValues,
				YValues: yValues,
			},
		},
	}

	err = graph.Render(chart.PNG, pngMetricsWriter)
	if err != nil {
		return err
	}

	return nil
}
