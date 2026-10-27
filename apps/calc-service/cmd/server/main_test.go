package main

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testContext() context.Context {
	return zerolog.New(io.Discard).WithContext(context.Background())
}

func receive[T any](t *testing.T, ch <-chan T) T {
	t.Helper()
	select {
	case value := <-ch:
		return value
	case <-time.After(3 * time.Second):
		require.FailNow(t, "timed out waiting for server activity")
	}
	var zero T
	return zero
}

func TestServeReturnsStartupError(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, listener.Close())
	})
	ctx, cancel := context.WithTimeout(testContext(), time.Second)
	defer cancel()
	err = serve(ctx, &http.Server{Addr: listener.Addr().String()})
	var networkError *net.OpError
	require.ErrorAs(t, err, &networkError, "serve must return a listener startup error")
}

func TestServeTreatsClosedServerAsNormal(t *testing.T) {
	srv := &http.Server{Addr: "127.0.0.1:0"}
	require.NoError(t, srv.Close())
	ctx, cancel := context.WithTimeout(testContext(), time.Second)
	defer cancel()
	require.NoError(t, serve(ctx, srv), "http.ErrServerClosed must not be a process failure")
}

func TestServeDrainsActiveRequestsAfterCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(testContext())
	defer cancel()
	listening := make(chan string, 1)
	entered := make(chan struct{})
	release := make(chan struct{})
	shutdownStarted := make(chan struct{})
	unblock := sync.OnceFunc(func() { close(release) })
	srv := &http.Server{
		Addr: "127.0.0.1:0",
		BaseContext: func(listener net.Listener) context.Context {
			listening <- listener.Addr().String()
			return context.Background()
		},
		Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			close(entered)
			<-release
			w.WriteHeader(http.StatusNoContent)
		}),
	}
	srv.RegisterOnShutdown(func() { close(shutdownStarted) })
	t.Cleanup(func() { _ = srv.Close() })
	t.Cleanup(unblock)
	stopped := make(chan error, 1)
	go func() { stopped <- serve(ctx, srv) }()
	address := receive(t, listening)
	requestDone := make(chan error, 1)
	go func() {
		client := &http.Client{Timeout: 3 * time.Second}
		response, err := client.Get("http://" + address)
		if err == nil {
			if response.StatusCode != http.StatusNoContent {
				err = errors.New("active request did not finish with 204")
			}
			err = errors.Join(err, response.Body.Close())
		}
		requestDone <- err
	}()
	receive(t, entered)
	cancel()
	receive(t, shutdownStarted)
	select {
	case err := <-stopped:
		require.FailNow(t, "serve returned before the active request finished", "error: %v", err)
	default:
	}
	unblock()
	require.NoError(t, receive(t, requestDone), "shutdown interrupted an active request")
	require.NoError(t, receive(t, stopped), "cancellation must use an independent shutdown context")
}

func TestGracefulShutdownForceClosesOnDeadline(t *testing.T) {
	entered := make(chan struct{})
	exited := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		close(entered)
		<-r.Context().Done()
		close(exited)
	}))
	t.Cleanup(srv.Close)
	requestDone := make(chan error, 1)
	go func() {
		response, err := srv.Client().Get(srv.URL)
		if err == nil {
			err = response.Body.Close()
		}
		requestDone <- err
	}()
	receive(t, entered)
	ctx, cancel := context.WithDeadline(testContext(), time.Now().Add(-time.Second))
	defer cancel()
	require.ErrorIs(t, gracefulShutdown(ctx, srv.Config), context.DeadlineExceeded)
	receive(t, exited)
	require.Error(t, receive(t, requestDone), "the unfinished connection must be forcibly closed after the deadline")
}
