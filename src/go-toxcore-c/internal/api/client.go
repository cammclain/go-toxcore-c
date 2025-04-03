package api

// Client is the interface for the Tox client API

// • Define a central type (e.g., ToxClient) that encapsulates the Tox client’s state and exposes methods for initialization, bootstrapping, messaging, and shutdown.

// • Use modern Go error handling (with %w for wrapping and, if needed, errors.Join for aggregated errors) and context (for cancellation and timeouts).

// Define API Types & Methods:
// • Create ToxClient with methods for:

// Initialization: Constructing a client using options.

// Bootstrap: Connecting to Tox bootstrap servers.

// Messaging: Sending and receiving text (and later, file) messages.

// Event Loop Management: Methods to run and stop the Tox iteration loop.

// Shutdown: Graceful termination of the client.
