package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
)

func newProvider() provider.Provider {
	return New("test")()
}

func TestProvider_Metadata(t *testing.T) {
	var resp provider.MetadataResponse
	newProvider().Metadata(context.Background(), provider.MetadataRequest{}, &resp)
	if resp.TypeName != "vpsie" {
		t.Errorf("expected type name vpsie, got %q", resp.TypeName)
	}
}

func TestProvider_Schema(t *testing.T) {
	var resp provider.SchemaResponse
	newProvider().Schema(context.Background(), provider.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %v", resp.Diagnostics)
	}

	at, ok := resp.Schema.Attributes["access_token"]
	if !ok {
		t.Fatal("expected access_token attribute")
	}
	if !at.IsSensitive() {
		t.Error("access_token must be marked sensitive")
	}
	if _, ok := resp.Schema.Attributes["endpoint"]; !ok {
		t.Error("expected endpoint attribute")
	}
}

func TestProvider_ResourcesAndDataSources(t *testing.T) {
	ctx := context.Background()
	p := newProvider()

	resources := p.Resources(ctx)
	if len(resources) < 20 {
		t.Errorf("expected many resources, got %d", len(resources))
	}
	for i, fn := range resources {
		if fn == nil || fn() == nil {
			t.Errorf("resource constructor %d produced nil", i)
		}
	}

	dataSources := p.DataSources(ctx)
	if len(dataSources) < 20 {
		t.Errorf("expected many data sources, got %d", len(dataSources))
	}
	for i, fn := range dataSources {
		if fn == nil || fn() == nil {
			t.Errorf("data source constructor %d produced nil", i)
		}
	}
}
