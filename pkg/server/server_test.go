package metrics

import (
	"fmt"
	"github.com/segmentio/stats/v4"
	"github.com/segmentio/stats/v4/prometheus"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

type MockHandler struct {
	Metrics     map[string]string
	RealHandler *prometheus.Handler
}

// stats.Handler implementation
func (h *MockHandler) HandleMeasures(time time.Time, measures ...stats.Measure) {
	h.RealHandler.HandleMeasures(time, measures...)
	for _, m := range measures {
		h.Metrics[m.Name] = m.String()
	}
}
func (h *MockHandler) Flush() {
	h.RealHandler.WriteStats(os.Stdout)
}

// http.Handler implementation
func (h *MockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.RealHandler.ServeHTTP(w, r)
}

func (h *MockHandler) Check(t *testing.T, name string, val string) {
	t.Helper()

	if m, ok := h.Metrics[name]; !ok {
		t.Errorf("metric not found: %s", name)
	} else if m != val {
		t.Errorf("metric unexpected val: exp(%s) act(%s)", val, m)
	}
}

type TestInstanceData struct {
	Eng         *stats.Engine
	Handler     *MockHandler
	StatsServer *Server
	HttpServer  *httptest.Server
}

func (t *TestInstanceData) Setup(statPrefix string) {
	// Create a new metrics router at port 9999
	t.Eng = stats.NewEngine(statPrefix, nil)
	t.Handler = &MockHandler{
		Metrics:     map[string]string{},
		RealHandler: &prometheus.Handler{},
	}
	t.StatsServer = NewPrometheusMetricServer(9999, t.Handler, t.Eng)

	// Attach that Handler to our httptest server
	t.HttpServer = httptest.NewServer(t.StatsServer)
}

func (t *TestInstanceData) TearDown() {
	t.HttpServer.Close()
	t.StatsServer.Close()
}

func TestServer_Run_metrics(t *testing.T) {
	td := TestInstanceData{}
	td.Setup("test_stats")
	defer td.TearDown()

	// Create a couple stats to work with
	now := time.Time{}
	td.Eng.IncrAt(now, "test.counter.A")
	td.Eng.IncrAt(now, "test.tagged.B", stats.Tag{
		Name:  "val1",
		Value: "testB",
	})

	for _, tc := range []struct {
		name      string
		fieldType string
		fieldName string
		fieldVal  string
		tags      string
	}{
		{
			"test_stats.test.counter",
			"counter",
			"A",
			"1",
			"[]",
		},
		{
			"test_stats.test.tagged",
			"counter",
			"B",
			"1",
			"[val1=testB]",
		},
	} {
		val := fmt.Sprintf("{ %s(%s:%s=%s) %s }", tc.name, tc.fieldType, tc.fieldName, tc.fieldVal, tc.tags)
		td.Handler.Check(t, tc.name, val)
	}
}

func TestServer_Run_results(t *testing.T) {
	td := TestInstanceData{}
	td.Setup("test_stats")
	defer td.TearDown()

	// Create a couple stats to work with
	now := time.Time{}
	td.Eng.IncrAt(now, "test.counter.A")
	td.Eng.IncrAt(now, "test.tagged.B")
	td.Eng.IncrAt(now, "test.tagged.B", stats.Tag{
		Name:  "val1",
		Value: "testB",
	})

	// Fetch the stats we just counted
	resp, err := http.Get(fmt.Sprintf("%s/metrics", td.HttpServer.URL))
	if err != nil {
		t.Fatalf("failed to fetch metrics")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	expected :=
		`# TYPE test_stats_test_counter_A counter
test_stats_test_counter_A 1

# TYPE test_stats_test_tagged_B counter
test_stats_test_tagged_B 1
test_stats_test_tagged_B{val1="testB"} 1
`

	if string(body) != expected {
		t.Errorf("unexpected value: exp(\n%s\n) act(\n%s\n)", expected, body)
	}
}
