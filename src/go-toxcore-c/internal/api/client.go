// Package api provides a high-level wrapper for the Tox protocol implementation.
// It abstracts the low-level C bindings to provide a more idiomatic Go interface.
package api

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
	// Import the tox package that provides the low-level bindings.
)

// ClientOptions wraps tox.ToxOptions for initializing a Tox client.
// This abstraction allows users to configure the Tox client without
// directly interacting with the low-level options structure.
type ClientOptions struct {
	// ToxOptions can be preconfigured before passing in.
	// If nil, NewClient will initialize a default set.
	// This allows advanced users to configure specific Tox options.
	ToxOptions *tox.ToxOptions
}

// BootstrapServer represents a Tox bootstrap server configuration.
// Bootstrap servers are used to connect to the Tox network initially.
// They provide entry points to the distributed network.
type BootstrapServer struct {
	// Address is the hostname or IP address of the bootstrap server
	Address string
	// Port is the UDP port of the bootstrap server
	Port uint16
	// PublicKey is the public key of the bootstrap server, used to verify its identity
	PublicKey string
}

// ToxClient encapsulates the Tox client's state and exposes API methods.
// It provides a high-level interface to interact with the Tox network,
// handling all the low-level details and lifecycle management.
type ToxClient struct {
	// client is the underlying Tox instance from the C library
	client *tox.Tox

	// ctx and cancel manage the event loop lifecycle.
	// These allow for graceful shutdown of background operations.
	ctx    context.Context
	cancel context.CancelFunc
	// wg is used to track the completion of the event loop goroutine
	wg sync.WaitGroup
}

// NewClient creates a new ToxClient instance using the provided options.
// It initializes the underlying Tox instance and sets up the context
// for managing the client's lifecycle.
//
// Parameters:
//   - options: Configuration options for initializing the Tox client
//
// Returns:
//   - A new ToxClient instance and nil if successful
//   - nil and an error if initialization fails
func NewClient(options *ClientOptions) (*ToxClient, error) {
	var opts *tox.ToxOptions
	// Use default options if none are provided
	if options == nil || options.ToxOptions == nil {
		opts = tox.NewToxOptions()
	} else {
		opts = options.ToxOptions
	}
	// Initialize the underlying Tox instance
	t := tox.NewTox(opts)
	if t == nil {
		return nil, errors.New("failed to initialize tox client")
	}
	// Create a cancellable context for managing the event loop
	ctx, cancel := context.WithCancel(context.Background())
	return &ToxClient{
		client: t,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// Bootstrap connects the Tox client to the specified bootstrap servers.
// This is necessary to join the Tox network by connecting to known nodes.
// The function attempts to connect to all provided servers and reports
// any errors encountered.
//
// Parameters:
//   - servers: A slice of BootstrapServer configurations to connect to
//
// Returns:
//   - nil if at least one server connection succeeds
//   - An error aggregating all connection failures if all attempts fail
func (c *ToxClient) Bootstrap(servers []BootstrapServer) error {
	var errs []error
	for _, server := range servers {
		if err := c.client.Bootstrap(server.Address, server.Port, server.PublicKey); err != nil {
			errs = append(errs, fmt.Errorf("failed to bootstrap with server %s: %w", server.Address, err))
		}
	}
	if len(errs) > 0 {
		// Aggregate errors (using errors.Join from Go 1.20+).
		return errors.Join(errs...)
	}
	return nil
}

// SendMessage sends a text message to a friend identified by friendNumber.
// The friendNumber is obtained when adding a friend or through friend-related callbacks.
//
// Parameters:
//   - friendNumber: The unique identifier for the target friend
//   - message: The text message to send
//
// Returns:
//   - nil if the message was sent successfully
//   - An error if sending the message failed
func (c *ToxClient) SendMessage(friendNumber uint32, message string) error {
	if _, err := c.client.FriendSendMessage(friendNumber, message); err != nil {
		return fmt.Errorf("failed to send message to friend %d: %w", friendNumber, err)
	}
	return nil
}

// RegisterMessageHandler registers a callback to handle incoming friend messages.
// The provided handler function will be called whenever a message is received from any friend.
//
// Parameters:
//   - handler: A function that processes incoming messages, receiving the friend number and message content
func (c *ToxClient) RegisterMessageHandler(handler func(friendNumber uint32, message string)) {
	c.client.CallbackFriendMessage(func(t *tox.Tox, friendNumber uint32, message string, userData interface{}) {
		handler(friendNumber, message)
	}, nil)
}

// StartEventLoop starts the Tox client's event loop with the specified iteration interval.
// The event loop processes network events and triggers callbacks at regular intervals.
// This method launches a background goroutine that continues until StopEventLoop is called.
//
// Parameters:
//   - iterationInterval: The time duration between consecutive event loop iterations
func (c *ToxClient) StartEventLoop(iterationInterval time.Duration) {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		ticker := time.NewTicker(iterationInterval)
		defer ticker.Stop()
		for {
			select {
			case <-c.ctx.Done():
				// Exit the loop when the context is cancelled
				return
			case <-ticker.C:
				// Process events at each tick interval
				c.client.Iterate()
			}
		}
	}()
}

// StopEventLoop stops the Tox client's event loop.
// This method cancels the context to signal the event loop to exit,
// then waits for the goroutine to complete.
func (c *ToxClient) StopEventLoop() {
	c.cancel()
	c.wg.Wait()
}

// Shutdown gracefully shuts down the Tox client.
// This method stops the event loop and cleans up the underlying Tox instance.
// It should be called when the client is no longer needed to prevent resource leaks.
//
// Returns:
//   - nil (currently always returns nil, but maintains the error return for future compatibility)
func (c *ToxClient) Shutdown() error {
	c.StopEventLoop()
	// Clean up the underlying tox instance.
	c.client.Kill()
	return nil
}
