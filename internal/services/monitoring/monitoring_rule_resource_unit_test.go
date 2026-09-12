package monitoring

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vpsie/govpsie"
)

func TestBuildMetricReqs(t *testing.T) {
	metrics := []monitoringMetricModel{
		{
			MetricType:    types.StringValue("cpu"),
			Condition:     types.StringValue("greater_than"),
			Threshold:     types.StringValue("80"),
			ThresholdType: types.StringValue("percentage"),
			Period:        types.StringValue("5"),
			Status:        types.StringNull(), // exercise the default
			Actions: []monitoringActionModel{
				{ActionName: types.StringValue("send_alert"), ActionKey: types.StringValue("send_alert"), Email: types.StringValue("a@b.c")},
			},
		},
	}

	reqs := buildMetricReqs(metrics)
	if len(reqs) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(reqs))
	}
	if reqs[0].MetricType != "cpu" || reqs[0].Threshold != "80" || reqs[0].Period != "5" {
		t.Errorf("unexpected metric: %+v", reqs[0])
	}
	if reqs[0].Status != "1" {
		t.Errorf("expected status to default to 1, got %q", reqs[0].Status)
	}
	if len(reqs[0].Actions) != 1 || reqs[0].Actions[0].ActionKey != "send_alert" {
		t.Errorf("unexpected actions: %+v", reqs[0].Actions)
	}
}

func TestMetricsFromAPI(t *testing.T) {
	metrics := []govpsie.MonitoringMetric{{
		MetricType: "ram", Condition: "less_than", Threshold: 20, ThresholdType: "percentage", Period: 10, Status: 1,
		Actions: []govpsie.MonitoringAction{{ActionName: "send_alert", ActionKey: "send_alert", Email: "x@y.z", Value: ""}},
	}}

	got := metricsFromAPI(metrics)
	if len(got) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(got))
	}
	if got[0].Threshold.ValueString() != "20" || got[0].Period.ValueString() != "10" {
		t.Errorf("int fields should become strings: %+v", got[0])
	}
	if len(got[0].Actions) != 1 {
		t.Fatalf("expected 1 action")
	}
	if !got[0].Actions[0].Value.IsNull() {
		t.Errorf("empty value should map to null, got %+v", got[0].Actions[0].Value)
	}
	if got[0].Actions[0].Email.ValueString() != "x@y.z" {
		t.Errorf("email should be preserved")
	}
}

func TestStringsDifference(t *testing.T) {
	diff := stringsDifference([]string{"a", "b", "c"}, []string{"b"})
	if len(diff) != 2 || diff[0] != "a" || diff[1] != "c" {
		t.Errorf("unexpected difference: %+v", diff)
	}
	if d := stringsDifference(nil, []string{"x"}); len(d) != 0 {
		t.Errorf("empty minus anything is empty, got %+v", d)
	}
}
