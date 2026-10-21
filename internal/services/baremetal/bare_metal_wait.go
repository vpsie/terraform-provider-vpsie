package baremetal

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/vpsie/govpsie"
)

// Timing of the waits. Each wait is bounded by the context deadline, which
// is the create, update or delete timeout of the resource; the counters
// below bound how long a wait goes on while the API keeps failing.
var (
	// pollInterval is the pause between two reads of a server or of its
	// private networks, and between two lookups of an unconfirmed order.
	pollInterval = 30 * time.Second
	// retryInterval is the pause before an order or a refused change is
	// sent again.
	retryInterval = 15 * time.Second
	// deployAnswerBound is the longest the API takes to answer an order: its
	// ingress timeout is 300 s and the deploy gives up by itself at 270 s.
	// One attempt is cut off after it. A 409 bareMetalDeployInProgress that
	// still comes back once this much time has passed since the first
	// attempt means the API sold the server but could not store its answer
	// (the key then stays in progress for 24 h), so the order is looked up
	// by hostname from then on.
	deployAnswerBound = 6 * time.Minute
	// orderLookupTimeout bounds the last lookup of an order whose outcome is
	// still unknown when the create timeout runs out, and the read of a
	// server found that way.
	orderLookupTimeout = time.Minute
)

const (
	// maxConsecutiveReadFailures bounds the reads in a row that may fail
	// (other than with 404) before a wait gives up.
	maxConsecutiveReadFailures = 5
	// maxConsecutiveNotFound bounds the reads in a row that may answer 404
	// while waiting for an order that was sold from stock: such a server is
	// readable at once, so a lasting 404 means it was taken back.
	maxConsecutiveNotFound = 3
	// minCloudflareStatus is the first of the 52x statuses Cloudflare answers
	// with when the origin fails or is too slow (524 after 100 s); the
	// request may still have reached the API.
	minCloudflareStatus = 520
)

// API error codes the resource reacts to.
const (
	codeDeployInProgress    = "bareMetalDeployInProgress"
	codeBusy                = "bareMetalBusy"
	codePrivateNetworkBusy  = "private_network_busy"
	codeNetworkNotAttached  = "private_network_not_attached"
	displayStateSuspended   = "suspended"
	releaseReasonDefault    = "Destroyed by Terraform"
	idempotencyKeyPrefix    = "terraform-"
	idempotencyKeyRandBytes = 16
)

// errServerGone is returned by a wait when the server stopped being readable.
var errServerGone = errors.New("the server is no longer on the account: the order was taken back or the server was released")

// sleep waits for d or until ctx ends.
func sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// timedOut wraps a context error with what was being waited for.
func timedOut(ctx context.Context, what string, last error) error {
	if last != nil {
		return fmt.Errorf("timed out waiting for %s (last error: %v): %w", what, last, ctx.Err())
	}

	return fmt.Errorf("timed out waiting for %s: %w", what, ctx.Err())
}

// newIdempotencyKey returns a random Idempotency-Key for one order.
func newIdempotencyKey() (string, error) {
	b := make([]byte, idempotencyKeyRandBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating the order's idempotency key: %w", err)
	}

	return idempotencyKeyPrefix + hex.EncodeToString(b), nil
}

// hasCode reports whether err is an API answer with the given status and code.
func hasCode(err error, status int, code string) bool {
	var apiErr *govpsie.APIError

	return errors.As(err, &apiErr) && apiErr.StatusCode == status && apiErr.Code == code
}

// orderOutcomeUnknown reports whether an answer to an order leaves open
// whether the server was sold, so the order is sent again with the same
// Idempotency-Key: no answer arrived (transport failure or the attempt was
// cut off), a gateway or Cloudflare failed (502, 503, 504, 52x), or the API
// still holds the key (409 bareMetalDeployInProgress). The API answers a
// resend of an order that went through with that order's answer, and runs a
// resend of an order that failed (the failure gave the key back) as a new
// order. A plain 500 or any other 4xx refuses this attempt. It says nothing
// about an earlier attempt, though: authentication, the permission check and
// the edge (WAF 403, 429) can answer before the API reads the key, so deploy
// looks the order up before it returns a refusal that follows an unknown
// outcome.
func orderOutcomeUnknown(err error) bool {
	var apiErr *govpsie.APIError
	if !errors.As(err, &apiErr) {
		return true
	}

	switch {
	case apiErr.StatusCode == http.StatusBadGateway,
		apiErr.StatusCode == http.StatusServiceUnavailable,
		apiErr.StatusCode == http.StatusGatewayTimeout,
		apiErr.StatusCode >= minCloudflareStatus:
		return true
	case apiErr.StatusCode == http.StatusConflict:
		return apiErr.Code == codeDeployInProgress
	default:
		return false
	}
}

// accountServers returns the identifiers of the account's servers, taken
// before an order so that a lookup after it can tell the new server from
// the existing ones. It returns nil when the list cannot be read; the order
// then goes ahead, but cannot be adopted by a lookup.
func accountServers(ctx context.Context, client *govpsie.Client) map[string]bool {
	servers, err := client.BareMetal.List(ctx)
	if err != nil {
		tflog.Warn(ctx, "could not list the bare-metal servers before the order; an unconfirmed order cannot be looked up",
			map[string]any{"error": err.Error()})
		return nil
	}

	ids := make(map[string]bool, len(servers))
	for _, s := range servers {
		ids[s.Identifier] = true
	}

	return ids
}

// findOrder looks for the server an unconfirmed order sold: a server in the
// order's datacenter, named with its hostname, that was not on the account
// before the order. It returns nil when there is none, and an error when
// the list cannot be read or the match is not unique.
func findOrder(ctx context.Context, client *govpsie.Client, req *govpsie.BareMetalDeployRequest, before map[string]bool) (*govpsie.BareMetalSummary, error) {
	if before == nil {
		return nil, errors.New("the account's servers could not be listed before the order, so a new server cannot be told from an existing one")
	}

	servers, err := client.BareMetal.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing the account's servers: %w", err)
	}

	var found []govpsie.BareMetalSummary
	for _, s := range servers {
		if !before[s.Identifier] && s.DatacenterIdentifier == req.DCIdentifier && strings.EqualFold(s.MachineName, req.Hostname) {
			found = append(found, s)
		}
	}

	switch len(found) {
	case 0:
		return nil, nil
	case 1:
		return &found[0], nil
	default:
		ids := make([]string, 0, len(found))
		for _, s := range found {
			ids = append(ids, s.Identifier)
		}
		return nil, fmt.Errorf("%d new servers are named %q in datacenter %s: %s", len(found), req.Hostname, req.DCIdentifier, strings.Join(ids, ", "))
	}
}

// adoptedOrder is the order answer of a server found by findOrder.
func adoptedOrder(s *govpsie.BareMetalSummary) *govpsie.BareMetalDeployment {
	return &govpsie.BareMetalDeployment{
		Success:    true,
		Identifier: s.Identifier,
		OSPhase:    s.OSPhase,
		IPAddress:  s.IPAddress,
		Status:     govpsie.BareMetalOrderSold,
	}
}

// deploy sends the order with one Idempotency-Key until the API answers it
// or ctx (the create timeout) ends. An answer that leaves the outcome open
// (orderOutcomeUnknown) is resent every retryInterval, so the number of
// attempts is bounded by the create timeout divided by retryInterval; one
// attempt lasts at most deployAnswerBound. Once the API has kept the key in
// progress for longer than deployAnswerBound, every round also looks the
// order up by hostname (findOrder), every pollInterval. When ctx ends with
// the outcome still open, one last lookup runs within orderLookupTimeout. A
// refusal ends the loop: it is returned as it is when every earlier attempt
// was refused or none was made, and after an unknown outcome it is returned
// only after one lookup within orderLookupTimeout. A server found by a
// lookup is returned with adopted set; when none is found the error names
// the hostname, so the user checks the account before ordering again.
func deploy(ctx context.Context, client *govpsie.Client, req *govpsie.BareMetalDeployRequest) (order *govpsie.BareMetalDeployment, adopted bool, err error) {
	key, err := newIdempotencyKey()
	if err != nil {
		return nil, false, err
	}

	before := accountServers(ctx, client)
	start := time.Now()
	var lastErr error
	// outcomeOpen records that an attempt of this order had an unknown
	// outcome, so that the server may have been sold.
	outcomeOpen := false
	for attempt := 1; ; attempt++ {
		if attempt > 1 {
			wait := retryInterval
			if hasCode(lastErr, http.StatusConflict, codeDeployInProgress) && time.Since(start) > deployAnswerBound {
				found, lookupErr := findOrder(ctx, client, req, before)
				if found != nil {
					return adoptedOrder(found), true, nil
				}
				if lookupErr != nil {
					tflog.Warn(ctx, "could not look up the unconfirmed bare-metal order", map[string]any{"error": lookupErr.Error()})
				}
				wait = pollInterval
			}
			tflog.Warn(ctx, "bare-metal order not confirmed, sending it again with the same idempotency key",
				map[string]any{"attempt": attempt, "error": lastErr.Error()})
			if sleep(ctx, wait) != nil {
				break
			}
		}

		attemptCtx, cancel := context.WithTimeout(ctx, deployAnswerBound)
		out, sendErr := client.BareMetal.Deploy(attemptCtx, req, key)
		cancel()
		if sendErr == nil {
			return out, false, nil
		}
		lastErr = sendErr
		if ctx.Err() != nil {
			break
		}
		if !orderOutcomeUnknown(sendErr) {
			if !outcomeOpen {
				return nil, false, sendErr
			}
			// An earlier attempt may have sold the server; this refusal may
			// have come before the API read the key.
			return lastLookup(ctx, client, req, before, "was refused after an earlier attempt whose outcome is unknown", sendErr)
		}
		outcomeOpen = true
	}

	return lastLookup(ctx, client, req, before, "was sent but not confirmed before the create timeout", lastErr)
}

// lastLookup runs the lookup that ends an order whose outcome stayed open:
// on a context of its own, so that it also runs after the create timeout or
// a cancel, within orderLookupTimeout. A server it finds is adopted;
// otherwise the error says what happened to the order, names the hostname
// and wraps lastErr.
func lastLookup(ctx context.Context, client *govpsie.Client, req *govpsie.BareMetalDeployRequest, before map[string]bool, what string, lastErr error) (*govpsie.BareMetalDeployment, bool, error) {
	lookupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), orderLookupTimeout)
	defer cancel()
	found, lookupErr := findOrder(lookupCtx, client, req, before)
	if found != nil {
		return adoptedOrder(found), true, nil
	}

	lookup := "no new server with that name was found"
	if lookupErr != nil {
		lookup = "the lookup failed: " + lookupErr.Error()
	}

	return nil, false, fmt.Errorf("the order of server %q in datacenter %s %s, "+
		"and %s. It may still have been sold: check the bare-metal servers of the account before applying again, and "+
		"import a server named %q with terraform import instead of ordering another (last error: %w)",
		req.Hostname, req.DCIdentifier, what, lookup, req.Hostname, lastErr)
}

// waitForOS reads the server until its OS is running: after an order or a
// password reset. A failed install or reset, a suspension, or a server that
// keeps answering 404 ends the wait with an error. The returned server is
// the last one read, also with an error when it was read.
func waitForOS(ctx context.Context, client *govpsie.Client, identifier, what string) (*govpsie.BareMetal, error) {
	notFound, failures := 0, 0
	var last *govpsie.BareMetal
	var lastErr error

	for {
		bm, err := client.BareMetal.Get(ctx, identifier)
		switch {
		case err == nil:
			notFound, failures, lastErr, last = 0, 0, nil, bm
			switch {
			case bm.OSPhase == govpsie.BareMetalOSPhaseRunning:
				return bm, nil
			case bm.OSPhase == govpsie.BareMetalOSPhaseFailed:
				return bm, fmt.Errorf("%s failed on the server%s", what, reasonSuffix(bm.DisplayStateReason))
			case bm.DisplayState == displayStateSuspended:
				return bm, fmt.Errorf("the server was suspended while waiting for %s%s", what, reasonSuffix(bm.DisplayStateReason))
			}
			tflog.Debug(ctx, "waiting for the bare-metal OS", map[string]any{"identifier": identifier, "os_phase": bm.OSPhase, "display_state": bm.DisplayState})
		case govpsie.IsNotFound(err):
			notFound++
			lastErr = err
			if notFound >= maxConsecutiveNotFound {
				return last, errServerGone
			}
		default:
			if ctx.Err() != nil {
				return last, timedOut(ctx, what, err)
			}
			failures++
			lastErr = err
			if failures >= maxConsecutiveReadFailures {
				return last, fmt.Errorf("reading the server failed %d times in a row: %w", failures, err)
			}
		}

		if err := sleep(ctx, pollInterval); err != nil {
			return last, timedOut(ctx, what, lastErr)
		}
	}
}

func reasonSuffix(reason string) string {
	if reason == "" {
		return ""
	}

	return ": " + reason
}

// retryWhileBusy runs op until the API stops refusing it as busy (409 with
// one of busyCodes), bounded by ctx.
func retryWhileBusy(ctx context.Context, what string, op func() error, busyCodes ...string) error {
	for {
		err := op()
		if err == nil {
			return nil
		}

		busy := false
		for _, code := range busyCodes {
			if hasCode(err, http.StatusConflict, code) {
				busy = true
				break
			}
		}
		if !busy {
			return err
		}

		tflog.Debug(ctx, "bare-metal server busy, retrying", map[string]any{"action": what, "error": err.Error()})
		if sleepErr := sleep(ctx, retryInterval); sleepErr != nil {
			return timedOut(ctx, "the server to accept "+what, err)
		}
	}
}

// waitForNetwork reads the server's private networks until networkID is
// attached (attach) or gone (detach). A failed change ends the wait with the
// API's error code.
func waitForNetwork(ctx context.Context, client *govpsie.Client, identifier string, networkID int64, attach bool) error {
	verb := "detach"
	if attach {
		verb = "attach"
	}
	what := fmt.Sprintf("private network %d to %s", networkID, verb)

	failures := 0
	var lastErr error
	for {
		list, err := client.BareMetal.ListPrivateNetworks(ctx, identifier)
		if err != nil {
			if ctx.Err() != nil {
				return timedOut(ctx, what, err)
			}
			failures++
			lastErr = err
			if failures >= maxConsecutiveReadFailures {
				return fmt.Errorf("reading the private networks failed %d times in a row: %w", failures, err)
			}
		} else {
			failures, lastErr = 0, nil
			item := findNetwork(list, networkID)
			switch {
			case item == nil && !attach:
				return nil
			case item == nil:
				return fmt.Errorf("private network %d is no longer listed on the server while attaching", networkID)
			case item.Status == govpsie.BareMetalNetworkFailed:
				return fmt.Errorf("private network %d could not %s (error code %q)", networkID, verb, item.ErrorCode)
			case attach && item.Status == govpsie.BareMetalNetworkAttached:
				return nil
			}
		}

		if err := sleep(ctx, pollInterval); err != nil {
			return timedOut(ctx, what, lastErr)
		}
	}
}

func findNetwork(list *govpsie.BareMetalPrivateNetworks, networkID int64) *govpsie.BareMetalPrivateNetwork {
	if list == nil {
		return nil
	}
	for i := range list.Items {
		if list.Items[i].NetworkID == networkID {
			return &list.Items[i]
		}
	}

	return nil
}

// attachNetwork attaches networkID and waits until it is attached.
func attachNetwork(ctx context.Context, client *govpsie.Client, identifier string, networkID int64) error {
	err := retryWhileBusy(ctx, fmt.Sprintf("the attach of private network %d", networkID), func() error {
		_, err := client.BareMetal.AttachPrivateNetwork(ctx, identifier, networkID)
		return err
	}, codePrivateNetworkBusy, codeBusy)
	if err != nil {
		return fmt.Errorf("attaching private network %d: %w", networkID, err)
	}

	return waitForNetwork(ctx, client, identifier, networkID, true)
}

// detachNetwork detaches networkID and waits until it is gone. A network
// that is not attached any more counts as detached.
func detachNetwork(ctx context.Context, client *govpsie.Client, identifier string, networkID int64) error {
	err := retryWhileBusy(ctx, fmt.Sprintf("the detach of private network %d", networkID), func() error {
		return client.BareMetal.DetachPrivateNetwork(ctx, identifier, networkID)
	}, codePrivateNetworkBusy, codeBusy)
	if err != nil {
		if hasCode(err, http.StatusNotFound, codeNetworkNotAttached) {
			return nil
		}
		return fmt.Errorf("detaching private network %d: %w", networkID, err)
	}

	return waitForNetwork(ctx, client, identifier, networkID, false)
}

// waitForRelease reads the server until it answers 404.
func waitForRelease(ctx context.Context, client *govpsie.Client, identifier string) error {
	failures := 0
	var lastErr error
	for {
		_, err := client.BareMetal.Get(ctx, identifier)
		switch {
		case govpsie.IsNotFound(err):
			return nil
		case err == nil:
			failures, lastErr = 0, nil
		default:
			if ctx.Err() != nil {
				return timedOut(ctx, "the server to be released", err)
			}
			failures++
			lastErr = err
			if failures >= maxConsecutiveReadFailures {
				return fmt.Errorf("reading the server failed %d times in a row: %w", failures, err)
			}
		}

		if err := sleep(ctx, pollInterval); err != nil {
			return timedOut(ctx, "the server to be released", lastErr)
		}
	}
}
