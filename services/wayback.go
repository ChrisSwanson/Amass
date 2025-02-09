// Copyright 2017 Jeff Foley. All rights reserved.
// Use of this source code is governed by Apache 2 LICENSE that can be found in the LICENSE file.

package services

import (
	"context"
	"fmt"
	"time"

	"github.com/chrisswanson/Amass/v3/config"
	"github.com/chrisswanson/Amass/v3/eventbus"
	"github.com/chrisswanson/Amass/v3/requests"
)

// Wayback is the Service that handles access to the Wayback data source.
type Wayback struct {
	BaseService

	SourceType string
	domain     string
	baseURL    string
}

// NewWayback returns he object initialized, but not yet started.
func NewWayback(sys System) *Wayback {
	w := &Wayback{
		SourceType: requests.ARCHIVE,
		domain:     "web.archive.org",
		baseURL:    "http://web.archive.org/web",
	}

	w.BaseService = *NewBaseService(w, "Wayback", sys)
	return w
}

// Type implements the Service interface.
func (w *Wayback) Type() string {
	return w.SourceType
}

// OnStart implements the Service interface.
func (w *Wayback) OnStart() error {
	w.BaseService.OnStart()

	w.SetRateLimit(time.Second)
	return nil
}

// OnDNSRequest implements the Service interface.
func (w *Wayback) OnDNSRequest(ctx context.Context, req *requests.DNSRequest) {
	cfg := ctx.Value(requests.ContextConfig).(*config.Config)
	bus := ctx.Value(requests.ContextEventBus).(*eventbus.EventBus)
	if cfg == nil || bus == nil {
		return
	}

	if req.Name == "" || req.Domain == "" {
		return
	}

	if !cfg.IsDomainInScope(req.Name) {
		return
	}

	w.CheckRateLimit()

	names, err := crawl(ctx, w.baseURL, w.domain, req.Name, req.Domain)
	if err != nil {
		bus.Publish(requests.LogTopic, fmt.Sprintf("%s: %v", w.String(), err))
		return
	}

	for _, name := range names {
		bus.Publish(requests.NewNameTopic, &requests.DNSRequest{
			Name:   cleanName(name),
			Domain: req.Domain,
			Tag:    w.SourceType,
			Source: w.String(),
		})
	}
}
