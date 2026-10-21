package baremetal

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/vpsie/govpsie"
)

// Timing of the waits. Each wait is bounded by the context deadline, which
// is the create, update or delete timeout of the resource; the counters
// below bound how long a wait goes on while the API keeps failing.
var (
	// pollInterval is the pause between two reads of a server or of its
	// private networks.
	pollInterval = 30 * time.Second
	// retryInterval is the pause before an order or a refused change is
	// sent again.
	retryInterval = 15 * time.Second
)

const (
	// maxDeployAttempts bounds how often one order is sent. Every attempt
	// carries the same Idempotency-Key, so the API answers a repeat of an
	// order that went through with that order instead of selling a second
	// server.
	maxDeployAttempts = 5
	// maxConsecutiveReadFailures bounds the reads in a row that may fail
	// (other than with 404) before a wait gives up.
	maxConsecutiveReadFailures = 5
	// maxConsecutiveNotFound bounds the reads in a row that may answer 404
	// while waiting for an order that was sold from stock: such a server is
	// readable at once, so a lasting 404 means it was taken back.
	maxConsecutiveNotFound = 3
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

// deployRetryable reports whether an order may be sent again with the same
// Idempotency-Key: no answer arrived (transport failure), a gateway failed
// (502/503/504), or the API still runs the first attempt of this key (409
// bareMetalDeployInProgress). A plain 500 or any 4xx is a definite refusal.
func deployRetryable(err error) bool {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	var apiErr *govpsie.APIError
	if !errors.As(err, &apiErr) {
		return true
	}

	switch apiErr.StatusCode {
	case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	case http.StatusConflict:
		return apiErr.Code == codeDeployInProgress
	default:
		return false
	}
}

// deploy sends the order, at most maxDeployAttempts times, all with one
// Idempotency-Key.
func deploy(ctx context.Context, client *govpsie.Client, req *govpsie.BareMetalDeployRequest) (*govpsie.BareMetalDeployment, error) {
	key, err := newIdempotencyKey()
	if err != nil {
		return nil, err
	}

	var lastErr error
	for attempt := 1; attempt <= maxDeployAttempts; attempt++ {
		if attempt > 1 {
			tflog.Warn(ctx, "bare-metal order not confirmed, sending it again with the same idempotency key",
				map[string]any{"attempt": attempt, "max_attempts": maxDeployAttempts, "error": lastErr.Error()})
			if err := sleep(ctx, retryInterval); err != nil {
				return nil, timedOut(ctx, "the order to be accepted", lastErr)
			}
		}

		out, err := client.BareMetal.Deploy(ctx, req, key)
		if err == nil {
			return out, nil
		}
		if !deployRetryable(err) {
			return nil, err
		}
		lastErr = err
	}

	return nil, fmt.Errorf("the order was not confirmed after %d attempts: %w", maxDeployAttempts, lastErr)
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
