package router

import (
	"reflect"
	"testing"

	"api-v2ray/internal/model"
)

func TestModelsReturnsSortedUniqueEnabledModels(t *testing.T) {
	svc := New(&model.Config{
		Upstreams: []model.Upstream{
			{ID: "u-b", Enabled: true, Models: []string{" z-model ", "gpt-4"}},
			{ID: "u-disabled", Enabled: false, Models: []string{"disabled-model"}},
			{ID: "u-a", Enabled: true, Models: []string{"gpt-4", "", "claude"}},
		},
	})

	got := svc.Models()
	want := []string{"claude", "gpt-4", "z-model"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected models: got %#v, want %#v", got, want)
	}
}