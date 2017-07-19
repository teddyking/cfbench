package datadog

import "github.com/teddyking/cfbench/bench"

type Point [2]int64

type MetricSeries struct {
	Metric string   `json:"metric"`         // phase.ShortName
	Points []Point  `json:"points"`         // single point for this run: time.Now(), duration
	Type   string   `json:"type"`           // gauge
	Host   string   `json:"host,omitempty"` // empty
	Tags   []string `json:"tags,omitempty"` // format key:value, e.g. sha:abcd   or version:1.19.0-rc22
}

type JsonResult struct {
	Series []MetricSeries `json:"series"`
}

func BuildJSONOutput(phases bench.Phases) JsonResult {
	result := JsonResult{}
	for _, phase := range phases {
		newSeries := MetricSeries{
			Metric: phase.ShortName,
			Points: []Point{
				Point{phase.StartTimestamp, int64(phase.Duration())},
			},
			Type: "gauge",
		}
		result.Series = append(result.Series, newSeries)
	}
	return result
}
