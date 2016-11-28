package stats

import (
	"reflect"
	"testing"
	"time"
)

func TestEngine(t *testing.T) {
	engine := NewEngine(EngineConfig{
		Prefix: "test",
		Tags:   []Tag{{"hello", "world"}},
	})
	defer engine.Close()

	now := time.Now()

	a := MakeCounter(engine, "A")
	b := MakeGauge(engine, "B")
	c := MakeCounter(engine, "C", Tag{"context", "test"})
	d := MakeTimer(engine, "D").StartAt(now)

	a.Incr()
	b.Set(2)
	c.Add(3)

	d.StampAt("lap", now.Add(1*time.Second))
	d.StampAt("lap", now.Add(2*time.Second))
	d.StopAt(now.Add(3 * time.Second))

	// Give a bit of time for the engine to update its state.
	time.Sleep(10 * time.Millisecond)

	metrics := engine.State()
	sortMetrics(metrics)

	expects := []Metric{
		Metric{
			Type:   CounterType,
			Key:    "test.A?hello=world",
			Name:   "test.A",
			Tags:   []Tag{{"hello", "world"}},
			Value:  1,
			Sample: 1,
		},
		Metric{
			Type:   GaugeType,
			Key:    "test.B?hello=world",
			Name:   "test.B",
			Tags:   []Tag{{"hello", "world"}},
			Value:  2,
			Sample: 1,
		},
		Metric{
			Type:   CounterType,
			Key:    "test.C?context=test&hello=world",
			Name:   "test.C",
			Tags:   []Tag{{"context", "test"}, {"hello", "world"}},
			Value:  3,
			Sample: 1,
		},
		Metric{
			Type:   HistogramType,
			Group:  "test.D?hello=world&stamp=lap",
			Key:    "test.D?hello=world&stamp=lap#0",
			Name:   "test.D",
			Tags:   []Tag{{"hello", "world"}, {"stamp", "lap"}},
			Value:  1,
			Sample: 1,
		},
		Metric{
			Type:   HistogramType,
			Group:  "test.D?hello=world&stamp=lap",
			Key:    "test.D?hello=world&stamp=lap#1",
			Name:   "test.D",
			Tags:   []Tag{{"hello", "world"}, {"stamp", "lap"}},
			Value:  1,
			Sample: 1,
		},
		Metric{
			Type:   HistogramType,
			Group:  "test.D?hello=world&stamp=total",
			Key:    "test.D?hello=world&stamp=total#0",
			Name:   "test.D",
			Tags:   []Tag{{"hello", "world"}, {"stamp", "total"}},
			Value:  1,
			Sample: 1,
		},
	}

	for i := range metrics {
		metrics[i].Time = time.Time{} // reset because we can't predict that value
	}

	if !reflect.DeepEqual(metrics, expects) {
		t.Error("bad engine state:")

		for i := range metrics {
			m := metrics[i]
			e := expects[i]

			if !reflect.DeepEqual(m, e) {
				t.Logf("unexpected metric at index %d:\n<<< %#v\n>>> %#v", i, m, e)
			}
		}
	}
}
