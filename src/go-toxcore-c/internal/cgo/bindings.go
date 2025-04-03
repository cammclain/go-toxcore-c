package cgo

// this file contains the cgo bindings for the toxcore library
// Toxcore is a C library, so we need to use cgo to call it
// toxcore can be found at https://github.com/TokTok/c-toxcore

//• Low-Level Bindings: Create a dedicated package (e.g., internal/cgo) that contains all cgo directives and bindings.

// Avoid direct allocation of opaque C structs.

// Use full struct definitions in the cgo preamble if necessary.

// Leverage runtime/cgo.Handle for safe pointer passing.
