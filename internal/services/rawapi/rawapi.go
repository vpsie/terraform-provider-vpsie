// Package rawapi calls VPSie customer routes that the govpsie client has no
// service for (Migrate, Protect, Control, Files) through its request and
// retry machinery.
package rawapi

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/vpsie/govpsie"
)

// API wraps the shared govpsie client.
type API struct {
	Client *govpsie.Client
}

// FromProviderData turns the provider's configure data into an API.
func FromProviderData(data any) (*API, error) {
	client, ok := data.(*govpsie.Client)
	if !ok {
		return nil, fmt.Errorf("expected *govpsie.Client, got %T", data)
	}
	return &API{Client: client}, nil
}

// Call sends one request. in is the JSON body (nil for none) and out receives
// the decoded answer (nil to ignore it).
func (a *API) Call(ctx context.Context, method, path string, in, out any) error {
	req, err := a.Client.NewRequest(ctx, method, path, in)
	if err != nil {
		return err
	}
	return a.Client.Do(ctx, req, out)
}

// Escape makes an identifier safe as one path segment.
func Escape(identifier string) string {
	return url.PathEscape(identifier)
}

// Page is the page size the list helpers ask for.
const Page = 100

// Paged walks a list route page by page until a short page, appending what
// fetch returns. fetch gets the path with limit and offset already added.
func Paged[T any](ctx context.Context, path string, fetch func(ctx context.Context, pagedPath string) ([]T, error)) ([]T, error) {
	var all []T
	sep := "?"
	if u, err := url.Parse(path); err == nil && u.RawQuery != "" {
		sep = "&"
	}
	for offset := 0; ; offset += Page {
		items, err := fetch(ctx, fmt.Sprintf("%s%slimit=%d&offset=%d", path, sep, Page, offset))
		if err != nil {
			return nil, err
		}
		all = append(all, items...)
		if len(items) < Page {
			return all, nil
		}
	}
}

// Method names kept here so callers need no net/http import.
const (
	Get    = http.MethodGet
	Post   = http.MethodPost
	Put    = http.MethodPut
	Patch  = http.MethodPatch
	Delete = http.MethodDelete
)
