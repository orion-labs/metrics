package metrics

import (
	"bytes"
	"fmt"
	"github.com/segmentio/stats/v4"
	"github.com/segmentio/stats/v4/prometheus"
	"net/http/httptest"
	"time"
)

func Example() {
	// Create a new metrics router at port 9999
	h := &prometheus.Handler{}
	e := stats.NewEngine("gobot_example", h,
		stats.T("name", "example"),
		stats.T("custom", "on all metrics, beware cardinality"))
	statSvr := NewPrometheusMetricServer(9999, h, e)
	svr := httptest.NewServer(statSvr)
	defer statSvr.Close()
	defer svr.Close()

	// In non-test environment start server via Run() method
	//runC := make(chan error)
	//go func() {
	//	defer close(runC)
	//	if err := statSvr.Run(); err != nil {
	//		runC <- err
	//	}
	//}()

	// Clock to drive repeatability
	// StampMilli = "Jan _2 15:04:05.000"
	t, err := time.Parse(time.StampMilli, "Oct 14 06:18:30.123")
	if err != nil {
		panic(fmt.Errorf("cannot parse timestamp: %v", err))
	}

	// Counters -->
	e.IncrAt(t, "stats are labels")
	e.AddAt(t, "stats are labels", 10)
	e.IncrAt(t, "stats can have extra tags too", stats.Tag{
		Name:  "val1",
		Value: "testB",
	})

	// Guages -->
	e.SetAt(t, "a guage has a fixed value", false)
	e.SetAt(t, "a guage only changes sometimes", true)

	// Histograms -->
	t2, err := time.Parse(time.StampMilli, "Oct 14 20:20:20.202")
	if err != nil {
		panic(fmt.Errorf("cannot parse timestamp: %v", err))
	}

	e.ObserveAt(t, "durations tracked this way", t2.Sub(t))

	// Force a stats flush
	e.Flush()

	buf := bytes.NewBuffer(nil)
	h.WriteStats(buf)

	fmt.Println(buf.String())
	// Output:
	// # TYPE gobot_example_a_guage_has_a_fixed_value gauge
	// gobot_example_a_guage_has_a_fixed_value{custom="on all metrics, beware cardinality",name="example"} 0 -62142399689877
	//
	// # TYPE gobot_example_a_guage_only_changes_sometimes gauge
	// gobot_example_a_guage_only_changes_sometimes{custom="on all metrics, beware cardinality",name="example"} 1 -62142399689877
	//
	// # TYPE gobot_example_durations_tracked_this_way histogram
	// gobot_example_durations_tracked_this_way_count{custom="on all metrics, beware cardinality",name="example"} 2 -62142399689877
	// gobot_example_durations_tracked_this_way_sum{custom="on all metrics, beware cardinality",name="example"} 101020.158 -62142399689877
	//
	// # TYPE gobot_example_stats_are_labels counter
	// gobot_example_stats_are_labels{custom="on all metrics, beware cardinality",name="example"} 22 -62142399689877
	//
	// # TYPE gobot_example_stats_can_have_extra_tags_too counter
	// gobot_example_stats_can_have_extra_tags_too{custom="on all metrics, beware cardinality",name="example",val1="testB"} 2 -62142399689877
}
