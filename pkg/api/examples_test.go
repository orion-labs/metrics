package api

import (
	"bytes"
	"fmt"
	"github.com/segmentio/stats/v4"
	"github.com/segmentio/stats/v4/prometheus"
	"net/http/httptest"
	"time"
)

func exampleStatsTracking() {
	// Clock to drive repeatability
	// StampMilli = "Jan _2 15:04:05.000"
	t, err := time.Parse(time.StampMilli, "Oct 14 06:18:30.123")
	if err != nil {
		panic(fmt.Errorf("cannot parse timestamp: %v", err))
	}

	t2, err := time.Parse(time.StampMilli, "Oct 14 20:20:20.202")
	if err != nil {
		panic(fmt.Errorf("cannot parse timestamp: %v", err))
	}

	exampleCounters(t)
	exampleGuages(t)
	exampleHistograms(t, t2)
}

func exampleCounters(t time.Time) {
	// Counters -->
	stats.IncrAt(t, "stats are labels")
	stats.AddAt(t, "stats are labels", 10)
	stats.IncrAt(t, "stats can have extra tags too", stats.Tag{
		Name:  "val1",
		Value: "testB",
	})
}

func exampleGuages(t time.Time) {
	// Guages -->
	stats.SetAt(t, "a guage has a fixed value", false)
	stats.SetAt(t, "a guage only changes sometimes", true)
}

func exampleHistograms(t, t2 time.Time) {
	stats.ObserveAt(t, "durations tracked this way", t2.Sub(t))
}

func Example_Default() {

	// default port is 7418; default mode is "debug"

	metrics := NewDefaultMetrics()
	statSvr := metrics.MetricsSvr
	svr := httptest.NewServer(statSvr)
	defer metrics.Close()
	defer svr.Close()
	defer func() {
		DefaultStatsHandler = &prometheus.Handler{}
	}()

	// In non-test environment start server via Run() method
	//runC := make(chan error)
	//go func() {
	//	defer close(runC)
	//	if err := statSvr.Run(); err != nil {
	//		runC <- err
	//	}
	//}()
	exampleStatsTracking()

	// Force a stats flush
	metrics.Flush()

	buf := bytes.NewBuffer(nil)
	metrics.StatsHandler.WriteStats(buf)

	fmt.Println(buf.String())
	// Output:
	//# TYPE a_guage_has_a_fixed_value gauge
	//a_guage_has_a_fixed_value 0 -62142399689877
	//
	//# TYPE a_guage_only_changes_sometimes gauge
	//a_guage_only_changes_sometimes 1 -62142399689877
	//
	//# TYPE durations_tracked_this_way histogram
	//durations_tracked_this_way_count 1 -62142399689877
	//durations_tracked_this_way_sum 50510.079 -62142399689877
	//
	//# TYPE stats_are_labels counter
	//stats_are_labels 11 -62142399689877
	//
	//# TYPE stats_can_have_extra_tags_too counter
	//stats_can_have_extra_tags_too{val1="testB"} 1 -62142399689877

}

func Example_WithLabels() {
	// Create a new metrics router at port 9999
	metrics := NewDefaultMetrics().WithTags([]stats.Tag{
		stats.T("name", "example"),
		stats.T("owner", "TJ"),
	})
	statSvr := metrics.MetricsSvr
	svr := httptest.NewServer(statSvr)
	defer metrics.Close()
	defer svr.Close()
	defer func() {
		DefaultStatsHandler = &prometheus.Handler{}
	}()

	// In non-test environment start server via Run() method
	//runC := make(chan error)
	//go func() {
	//	defer close(runC)
	//	if err := statSvr.Run(); err != nil {
	//		runC <- err
	//	}
	//}()
	exampleStatsTracking()

	// Force a stats flush
	metrics.Flush()

	buf := bytes.NewBuffer(nil)
	metrics.StatsHandler.WriteStats(buf)

	fmt.Println(buf.String())
	//# TYPE a_guage_has_a_fixed_value gauge
	//a_guage_has_a_fixed_value{name="example",owner="TJ"} 0 -62142399689877
	//
	//# TYPE a_guage_only_changes_sometimes gauge
	//a_guage_only_changes_sometimes{name="example",owner="TJ"} 1 -62142399689877
	//
	//# TYPE durations_tracked_this_way histogram
	//durations_tracked_this_way_count{name="example",owner="TJ"} 1 -62142399689877
	//durations_tracked_this_way_sum{name="example",owner="TJ"} 50510.079 -62142399689877
	//
	//# TYPE stats_are_labels counter
	//stats_are_labels{name="example",owner="TJ"} 11 -62142399689877
	//
	//# TYPE stats_can_have_extra_tags_too counter
	//stats_can_have_extra_tags_too{name="example",owner="TJ",val1="testB"} 1 -62142399689877
}

func Example_Customized() {

	// customize port and mode to 9000 and "release"
	statsPort := 9000
	ginMode := "release"

	metrics := NewMetrics(statsPort, ginMode, &prometheus.Handler{},
		NewStatsEngine("example_with_prefix",
			stats.T("name", "example"),
			stats.T("owner", "TJ"),
			stats.T("custom", "on all metrics, beware cardinality")))
	statSvr := metrics.MetricsSvr
	svr := httptest.NewServer(statSvr)
	defer metrics.Close()
	defer svr.Close()

	// track some stats
	exampleStatsTracking()

	// Force a stats flush
	metrics.Flush()

	buf := bytes.NewBuffer(nil)
	metrics.StatsHandler.WriteStats(buf)

	fmt.Println(buf.String())
	// Output:
	//# TYPE example_with_prefix_a_guage_has_a_fixed_value gauge
	//example_with_prefix_a_guage_has_a_fixed_value{custom="on all metrics, beware cardinality",name="example",owner="TJ"} 0 -62142399689877
	//
	//# TYPE example_with_prefix_a_guage_only_changes_sometimes gauge
	//example_with_prefix_a_guage_only_changes_sometimes{custom="on all metrics, beware cardinality",name="example",owner="TJ"} 1 -62142399689877
	//
	//# TYPE example_with_prefix_durations_tracked_this_way histogram
	//example_with_prefix_durations_tracked_this_way_count{custom="on all metrics, beware cardinality",name="example",owner="TJ"} 1 -62142399689877
	//example_with_prefix_durations_tracked_this_way_sum{custom="on all metrics, beware cardinality",name="example",owner="TJ"} 50510.079 -62142399689877
	//
	//# TYPE example_with_prefix_stats_are_labels counter
	//example_with_prefix_stats_are_labels{custom="on all metrics, beware cardinality",name="example",owner="TJ"} 11 -62142399689877
	//
	//# TYPE example_with_prefix_stats_can_have_extra_tags_too counter
	//example_with_prefix_stats_can_have_extra_tags_too{custom="on all metrics, beware cardinality",name="example",owner="TJ",val1="testB"} 1 -62142399689877

}
