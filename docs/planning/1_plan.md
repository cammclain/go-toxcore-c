
# Upgrading Go Toxcore C to Version 1.23

I am currently interested in using the [Go Toxcore C](https://github.com/TokTok/go-toxcore-c) library and would like to upgrade to version 1.23. 

As it stands, go-toxcore-c is at Go v1.12 which seems very old.

So, I would like to know how to upgrade the library to version 1.23, for use with the rest of my Go project.


## Understanding go-toxcore-c

The `go-toxcore-c` library is a Go wrapper around the Toxcore C library, which provides the core functionality of the Tox protocol. 


# Understanding the changes from 1.12 to 1.23

Upgrading a Go cgo wrapper from Go 1.12 to Go 1.23.1 spans many Go releases. Below is a comprehensive summary of significant changes – including breaking changes and important new features – that may impact the **go-toxcore-c** wrapper. These are organized by category for clarity.

## Module system and build tool changes

- **Go modules become the default (Go 1.13–1.16):** Starting in Go 1.13, module support was greatly improved and a proxy (`proxy.golang.org`) was enabled by default ([What's new in Go 1.13](https://blog.kowalczyk.info/article/715712478e8d4dbab6b7c140c8f5e76e/whats-new-in-go-1.13.html#:~:text=go%20mod)). By Go 1.14, modules were considered production-ready and users were encouraged to migrate off GOPATH to modules ([Go 1.14 is released - The Go Programming Language](https://go.dev/blog/go1.14#:~:text=Some%20of%20the%20highlights%20include%3A)). In Go 1.16, module-aware mode was **enabled by default**, even without a `go.mod` (effectively `GO111MODULE=on` by default) ([Go 1.16 Release Notes - The Go Programming Language](https://go.dev/doc/go1.16#:~:text=Module,auto)). This means the project should have a `go.mod` and rely on modules for dependency management. Ensure the `go.mod` file’s `go` version is bumped appropriately and run `go mod tidy` under a newer Go toolchain to align dependencies.

- **`go get` and installing tools:** In module mode, the role of `go get` changed. Go 1.17 deprecated using `go get` to install binaries; instead `go install pkg@version` is now the recommended approach ([Go 1.16 Release Notes - The Go Programming Language](https://go.dev/doc/go1.16#:~:text=,dependencies%20of%20the%20main%20module)). By Go 1.22, `go get` is **no longer supported in GOPATH mode** (with `GO111MODULE=off`); all usage of `go get` now requires modules ([Go 1.22 Release Notes - The Go Programming Language](https://tip.golang.org/doc/go1.22#:~:text=,programs)). When upgrading, remove any old scripts or Dockerfiles that use `go get` in GOPATH mode for dependencies. Use `go mod tidy` to manage deps and `go install ...@latest` to install any tool binaries.

- **Vendoring changes:** If the project vendors dependencies, note that Go 1.17+ records the module versions in `vendor/modules.txt` and omits certain files for consistency ([Go 1.17 Release Notes - The Go Programming Language](https://go.dev/doc/go1.17#:~:text=)). Also, Go 1.22 introduced `go work vendor` for multi-module workspaces ([Go 1.22 Release Notes - The Go Programming Language](https://tip.golang.org/doc/go1.22#:~:text=Go%20command)). Generally, ensure your vendoring (if used) is regenerated with the new Go version to avoid surprises.

- **Build flags and tooling:** A few new build options were added that could aid the build process. For example, Go 1.13 added `-trimpath` to produce reproducible builds by omitting local file system paths in binaries ([What's new in Go 1.13](https://blog.kowalczyk.info/article/715712478e8d4dbab6b7c140c8f5e76e/whats-new-in-go-1.13.html#:~:text=Trim%20file%20paths%20recorded%20in,the%20binary)). Go 1.18 introduced `-asan` to integrate with C/C++ address sanitizer builds ([Go 1.18 Release Notes - The Go Programming Language](https://tip.golang.org/doc/go1.18#:~:text=%23%20%60go%60%20%60build%60%20%60)) (useful if you want to detect memory errors when linking against sanitizers). Also, the Go 1.20 toolchain started embedding version control info (VCS hash, time) in builds by default for `go install`-ed binaries; if deterministic builds are needed (e.g. in Bazel), you can use `-buildvcs=false` to disable that. 

- **Bazel build compatibility:** If you build with **Bazel**, ensure you update to a rules_go version that supports the newer Go releases. Bazel’s Go rules have evolved to handle module mode and new Go features (for example, supporting the `//go:embed` directive added in Go 1.16 ([support go:embed directives in 1.16 · Issue #2775 · bazel ... - GitHub](https://github.com/bazelbuild/rules_go/issues/2775#:~:text=support%20go%3Aembed%20directives%20in%201,go_binary%20%2C%20go_test))). After upgrading Go, bump the rules_go version accordingly. Also, because Go 1.18+ defaults to embedding build info, you might need to pass flags in your Bazel configuration to maintain reproducibility. In short, use an updated toolchain and expect faster, more consistent builds with the new Go toolchain.

## Changes in cgo behavior and FFI support

- **Stricter rules for C struct handling:** As of Go 1.15.3, cgo **forbids** Go code from directly allocating C structs that are incomplete (opaque) in C ([Go 1.15 Release Notes - The Go Programming Language](https://go.dev/doc/go1.15#:~:text=In%20Go%201,the%20appropriate%20C%20header%20file)). In practice, this means if the C library defines a struct only as `struct Tox;` (forward declaration) and expects you to use it via pointer, you can no longer do something like `new(C.Tox)` in Go. You must use pointers and have the full definition in the cgo preamble if you need to allocate it in Go. Review the wrapper to ensure you aren’t instantiating any C struct types directly – use C factory functions or maintain them as pointers. This change fixes unsafe behavior and might require minor code adjustments in how you obtain and use structures from the toxcore C library.

- **C struct bitfields no longer auto-mapped:** In Go 1.16, the cgo tool stopped trying to generate Go struct fields for C bitfields ([Go 1.16 Release Notes - The Go Programming Language](https://go.dev/doc/go1.16#:~:text=Cgo)). Previously, if toxcore had bitfields in its C structs, cgo might have created corresponding fields in the Go struct, but this was unreliable. Now those bitfields will be ignored, and the Go binding would need to manually handle any needed flags via access functions. Check if toxcore’s headers use bitfields; if so, ensure the wrapper isn’t assuming the old cgo behavior (likely not, unless you saw Go fields that suddenly disappear when upgrading).

- **Safe passing of Go pointers to C:** Go’s rules about passing Go pointers into C were already strict (you generally cannot pass a Go pointer to Go-allocated memory into C unless certain conditions are met). While the core rules remain, Go 1.17 added the `runtime/cgo.Handle` mechanism ([Go 1.17 Release Notes - The Go Programming Language](https://go.dev/doc/go1.17#:~:text=Cgo)) to safely create an opaque handle for any Go value that can be passed through C and back. If your wrapper was using unsafe workarounds to store Go callback references or contexts in C, you can now use `cgo.Handle` for a safer, idiomatic solution.

- **Linker flags handling:** For projects with many C linker flags, Go 1.23 improves cgo by supporting an `-ldflags` flag to pass options to the C linker ([Go 1.23 Release Notes - The Go Programming Language](https://tip.golang.org/doc/go1.23#:~:text=Cgo)). The Go command uses this automatically to avoid very long command lines. This isn’t a breaking change, but if your build was hitting “argument list too long” errors when linking (due to many `#cgo LDFLAGS` entries for toxcore libraries), this should be resolved in Go 1.23.

- **Cross-compilation and toolchain detection:** Go 1.20 changed the default `CGO_ENABLED` behavior: if no C compiler is found, cgo will default to off instead of erroring ([Go 1.20 Release Notes - The Go Programming Language](https://tip.golang.org/doc/go1.20#:~:text=Cgo)). This means on systems without a C toolchain (or in minimal containers), Go will silently build the pure-Go parts only. In context of the wrapper, you’ll want to ensure a C compiler is available (e.g., gcc/clang or even using Zig as a linker) when building, otherwise cgo might be disabled unintentionally. On macOS, the `net` and `os/user` std packages were internally rewritten to not require cgo ([Go 1.20 Release Notes - The Go Programming Language](https://tip.golang.org/doc/go1.20#:~:text=The%20packages%20in%20the%20standard,Go%20version%20of%20these%20packages)), which is mainly informative (makes pure-Go builds easier). Just be mindful to install a C compiler in build environments, or explicitly set `CGO_ENABLED=1`, when the toxcore C library needs to be linked.

*(In summary, audit the cgo preamble and usage: avoid doing anything now disallowed, like creating opaque C structs on the Go side, and take advantage of new features for safer pointer handling and smoother linking.)*

## Standard library updates and deprecations

- **`io/ioutil` deprecation (Go 1.16):** The `io/ioutil` package was marked deprecated ([Go 1.16 Release Notes - The Go Programming Language](https://go.dev/doc/go1.16#:~:text=Deprecation%20of%20io%2Fioutil)). All of its functions were **migrated to other packages**: e.g. `ioutil.ReadFile` -> `os.ReadFile`, `ioutil.WriteFile` -> `os.WriteFile`, `ioutil.TempDir` -> `os.MkdirTemp`, `ioutil.ReadAll` -> `io.ReadAll`, etc ([Go 1.16 Release Notes - The Go Programming Language](https://go.dev/doc/go1.16#:~:text=,70)). The old functions still work (and will continue to in Go 1.23), but you’ll see deprecation warnings. It’s advisable to update the wrapper’s code to use the new API locations for clarity and future-proofing. For instance, if `tox_save` was using `ioutil.WriteFile`, switch to `os.WriteFile` (same functionality).

- **New file embedding capability:** Go 1.16 introduced the `//go:embed` directive and an `embed` package for embedding static files into binaries ([Go 1.16 Release Notes - The Go Programming Language](https://go.dev/doc/go1.16#:~:text=)). If the toxcore wrapper or its tests include any static test data or resources, you can now embed them instead of reading from disk at runtime. This isn’t required but could simplify packaging (no need to ship separate resource files).

- **Context and cancellation:** In Go 1.20, the `context` package added `WithCancelCause` and `context.Cause` to allow attaching an error cause when canceling a context ([[持续更新ING]Go语言历史版本演进和新特性_goland历史版本- Ius7](https://www.ius7.com/a/434#:~:text=%5B%E6%8C%81%E7%BB%AD%E6%9B%B4%E6%96%B0ING%5DGo%E8%AF%AD%E8%A8%80%E5%8E%86%E5%8F%B2%E7%89%88%E6%9C%AC%E6%BC%94%E8%BF%9B%E5%92%8C%E6%96%B0%E7%89%B9%E6%80%A7_goland%E5%8E%86%E5%8F%B2%E7%89%88%E6%9C%AC,WithCancelCause%20%E6%8F%90%E4%BE%9B%E4%BA%86%E4%B8%80%E7%A7%8D%E6%96%B9%E6%B3%95%E6%9D%A5%E5%8F%96%E6%B6%88%E5%85%B7%E6%9C%89%E7%BB%99%E5%AE%9A%E9%94%99%E8%AF%AF%E7%9A%84%E4%B8%8A%E4%B8%8B%E6%96%87%3B%20os)). If the wrapper uses `context.Context` (for example, if any tox operations are cancellable), you can now propagate error reasons on cancellation. Additionally, Go 1.15 made it illegal to create a context from a nil parent (will panic) ([Go 1.15 Release Notes - The Go Programming Language](https://go.dev/doc/go1.15#:~:text=)) – a minor gotcha if any code was doing `context.WithCancel(nil)`. Use `context.Background()` as the base instead.

- **Testing improvements:** If the project’s tests are being updated, note that Go 1.14 added `t.Cleanup` in the `testing` package to register cleanup callbacks at test end (a nicer alternative to defer in tests) ([Golang Weekly Issue 300: February 21, 2020](https://golangweekly.com/issues/300#:~:text=What%27s%20New%20In%20Go%201,This%20is%20good)). Go 1.18 added built-in fuzz testing support (`go test -fuzz`), which could be interesting for fuzzing the wrapper’s C-Go interactions. And Go 1.19+ added `T.Setenv` for managing env vars in tests, etc. These changes don’t affect production code but can improve the test suite when you opt to use them.

- **Miscellaneous library changes:** Many other standard library improvements accumulated, generally maintaining backward compatibility. For example, random number generation (`math/rand`) is now seeded automatically (no need for `rand.Seed(time.Now().UnixNano())` in init) ([Go 1.20 in a nutshell](https://appliedgo.com/blog/go-1-20-in-a-nutshell#:~:text=)). The `crypto/x509` CommonName field handling was deprecated in 1.15 (now SANs are required, which could matter if toxcore deals with certificates) ([Go 1.15 Release Notes - The Go Programming Language](https://go.dev/doc/go1.15#:~:text=X)). Go 1.21 introduced new generic packages `slices`, `maps`, and `cmp` in the standard library for easier handling of slices and maps (useful if you decide to utilize generics in the wrapper code) ([Fourteen Years of Go - The Go Programming Language](https://go.dev/blog/14years#:~:text=Go%201,the%20blog%20post%20%E2%80%9C%2033%E2%80%9D)). Go 1.23 added `unique` (for value interning), `iter` (iterator building blocks), and `structs` (for struct field metadata) packages ([Go 1.23 Release Notes - The Go Programming Language](https://tip.golang.org/doc/go1.23#:~:text=New%20unique%20package)) ([Go 1.23 Release Notes - The Go Programming Language](https://tip.golang.org/doc/go1.23#:~:text=New%20structs%20package)) – these are advanced features that likely don’t affect the wrapper unless you plan to refactor using them. The key takeaway is that the standard library has grown; you might find new helpers and should be aware of deprecations when updating the code.

## Error handling enhancements

Go’s error handling patterns improved significantly since 1.12, which can inform how you handle errors in the wrapper:

- **Error wrapping with `%w` (Go 1.13):** Go 1.13 introduced a standardized way to wrap errors. The `fmt.Errorf` function now supports the `%w` verb to wrap an inner error ([What's new in Go 1.13](https://blog.kowalczyk.info/article/715712478e8d4dbab6b7c140c8f5e76e/whats-new-in-go-1.13.html#:~:text=To%20create%20an%20error%20that,if%20nothing%20is%20wrapped)). Correspondingly, the `errors` package gained `errors.Is`, `errors.As`, and `errors.Unwrap` to check error chains ([What's new in Go 1.13](https://blog.kowalczyk.info/article/715712478e8d4dbab6b7c140c8f5e76e/whats-new-in-go-1.13.html#:~:text=Error%20wrapping)). This means you can wrap errors returned from toxcore (for example, include context like “failed to initialize tox: …”) while still allowing callers to extract the underlying error. If the wrapper currently returns errors in custom ways, consider adopting this pattern for consistency. For instance, `return fmt.Errorf("tox new failed: %w", err)` will let callers do `errors.Is(err, someToxError)`.

- **Multiple errors and joining (Go 1.20):** Go 1.20 added `errors.Join`, which creates an error that wraps multiple errors at once ([Go 1.20 Release Notes - The Go Programming Language](https://tip.golang.org/doc/go1.20#:~:text=The%20new%20function%20errors,wrapping%20a%20list%20of%20errors)). It also updated `errors.Is/As` to traverse these multi-error chains. This isn’t likely a core need for go-toxcore-c, but if toxcore returns multiple error states or if you want to accumulate errors (say, closing multiple resources), the standard library now supports that. It’s a new option instead of crafting custom error types for aggregate errors.

- **Standard library now wraps errors:** Along with the above features, many standard library functions started returning wrapped errors to preserve more context. For example, in Go 1.23 database/sql began wrapping driver errors ([Go 1.23 Release Notes - The Go Programming Language](https://tip.golang.org/doc/go1.23#:~:text=)). As you upgrade, be mindful that error values may format differently (including additional text). Generally, your code doesn’t need changes for this; just use `errors.Is` or `errors.As` rather than string-matching error text, to be robust against these changes.

Overall, the upgrade allows more idiomatic error handling. You might refactor any ad-hoc error wrapping in go-toxcore-c to use `fmt.Errorf("%w", err)` and friends, making the API nicer for users without breaking compatibility (the Go 1.x compatibility promise means old error behaviors remain, but using the new features is purely additive).

## Compiler, runtime, and performance improvements

Many internal improvements from Go 1.13 through 1.23 will benefit the wrapper without requiring code changes, but they’re worth understanding:

- **Garbage collector and scheduler:** Go 1.14 introduced **asynchronous preemption** of goroutines, meaning loops and C calls can be preempted more smoothly for GC or scheduling ([Go 1.14 is released - The Go Programming Language](https://go.dev/blog/go1.14#:~:text=,Internal%20timers%20are%20more%20efficient)). This fixes issues where tight CPU loops could starve other goroutines. In practice, if toxcore C functions call into long loops or computations, the Go scheduler can now preempt them more safely, improving application responsiveness. Go 1.19 added a soft memory limit for the runtime heap (via `GOMEMLIMIT`) ([Go1.19 那些事：国产芯片、内存模型等新特性，你知道多少？](https://zhuanlan.zhihu.com/p/536274333#:~:text=%E4%BB%8A%E5%A4%A9%E5%B0%B1%E7%94%B1%E7%85%8E%E9%B1%BC%E5%92%8C%E5%A4%A7%E5%AE%B6%E5%9B%B4%E8%A7%82%E3%80%8AGo%201.19%20Release%20Notes,Cox%20%E5%86%99%E4%BA%86Go%20Memory%20Model%20%E7%9A%84%E4%B8%89%E7%AF%87%E6%96%87%E7%AB%A0%E4%BD%9C%E4%B8%BA%E7%B3%BB%E5%88%97%E8%AF%B4%E6%98%8E%EF%BC%9A)), allowing apps to trade performance for a capped memory usage – possibly useful in long-running tox applications to avoid uncontrolled memory growth. Each version has tuned the garbage collector and scheduler for lower latency and better parallelism (e.g., Go 1.20 reorganized GC structures for efficiency ([Go 1.20 Release Notes - The Go Programming Language](https://tip.golang.org/doc/go1.20#:~:text=Some%20of%20the%20garbage%20collector%E2%80%99s,change%20reduces%20memory%20overheads%20and))).

- **Compiler optimizations:** The compiler has gotten **smarter and faster**. Escape analysis was improved in Go 1.13 to put more objects on the stack (reducing GC load) ([What's new in Go 1.13](https://blog.kowalczyk.info/article/715712478e8d4dbab6b7c140c8f5e76e/whats-new-in-go-1.13.html#:~:text=New%20escape%20analysis)). Go 1.14–1.15 focused on faster defer (1.14) and huge linker speed improvements in 1.15 ([Go 1.15 Release Notes - The Go Programming Language](https://go.dev/doc/go1.15#:~:text=Go%201,tzdata%20package%20has%20been%20added)). Notably, Go 1.17 switched to a register-based calling convention on x86-64, which can yield CPU performance boosts for function calls ([Go 1.17 Release Notes - Hacker News](https://news.ycombinator.com/item?id=28201732#:~:text=Go%201.17%20Release%20Notes%20,bound%20programs)). For a cgo wrapper, the overhead of calling C functions remains roughly the same (cgo calls still involve some overhead), but Go->Go calls in your wrapper will be a bit faster. The linker in Go 1.17+ is much faster and produces smaller binaries due to dead-code elimination and other changes ([Go 1.15 Release Notes - The Go Programming Language](https://go.dev/doc/go1.15#:~:text=Go%201,tzdata%20package%20has%20been%20added)). All of this means your builds will be quicker and the resulting binary likely a bit leaner and faster at runtime, without any action needed on your part.

- **Improved concurrency and sync:** The `sync` package primitives and scheduler have been refined. For example, as of Go 1.13, a single deferred function is cheaper (one defer per function now uses stack allocation) ([What's new in Go 1.13](https://blog.kowalczyk.info/article/715712478e8d4dbab6b7c140c8f5e76e/whats-new-in-go-1.13.html#:~:text=Faster%20)). Go 1.18 introduced the new `runtime/metrics` package for more fine-grained telemetry of the runtime if needed ([Go 1.16 Release Notes - The Go Programming Language](https://go.dev/doc/go1.16#:~:text=Runtime)). Go 1.21 added `sync/atomic` types like `atomic.Int64` that simplify atomic usage (no more `atomic.LoadInt64(&x)` boilerplate if you choose to use them). These aren’t breaking changes, but if the wrapper has concurrency (perhaps event callbacks, background threads interfacing with toxcore), you may find it easier to implement with these improvements. Also, Go’s race detector no longer requires cgo on macOS as of Go 1.20 ([Go 1.20 Release Notes - The Go Programming Language](https://tip.golang.org/doc/go1.20#:~:text=requires%20passing%20%60,the%20C%20code)), so you can run `-race` on macOS without extra fuss.

- **New language features (optional to use):** Although not necessary for upgrading, be aware Go 1.18 introduced **Generics** (type parameters) into the language ([Go 1.18 Release Notes - The Go Programming Language](https://tip.golang.org/doc/go1.18#:~:text=Generics)). The go-toxcore-c codebase can remain as is – you don’t need to use generics – but you now *can* write type-safe generic functions or data structures if it benefits the wrapper (for example, one could imagine a generic helper for converting Go slices to C arrays). Similarly, Go 1.21 added new built-in functions `min`, `max`, and `clear` for convenience, and Go 1.22 even tweaked the language spec so that loop variables are redefined each iteration (fixing the long-standing “closure over loop variable” gotcha) ([Go 1.22 Release Notes - The Go Programming Language](https://tip.golang.org/doc/go1.22#:~:text=Go%201,to%20%E2%80%9Cfor%E2%80%9D%20loops)). None of these require changes to existing code, but they pave the way for cleaner code if you refactor or extend the wrapper.

In summary, each Go release from 1.13 to 1.23 has maintained backward compatibility with older code while introducing performance gains and better tooling. When upgrading the go-toxcore-c wrapper, the primary adjustments will be around adopting the Go modules workflow, adjusting any cgo usages that violate newer rules, and updating imports away from deprecated APIs like `io/ioutil`. Embracing the new error handling and other idioms can be done gradually but will make the wrapper more idiomatic. After these changes and a thorough re-test, the project should build and run successfully on Go 1.23.1, benefitting from the years of improvements in speed, safety, and maintainability. 

**Sources:**

- Go 1.13–1.23 official release notes ([What's new in Go 1.13](https://blog.kowalczyk.info/article/715712478e8d4dbab6b7c140c8f5e76e/whats-new-in-go-1.13.html#:~:text=go%20mod)) ([Go 1.14 is released - The Go Programming Language](https://go.dev/blog/go1.14#:~:text=Some%20of%20the%20highlights%20include%3A)) ([Go 1.15 Release Notes - The Go Programming Language](https://go.dev/doc/go1.15#:~:text=In%20Go%201,the%20appropriate%20C%20header%20file)) ([Go 1.16 Release Notes - The Go Programming Language](https://go.dev/doc/go1.16#:~:text=Cgo)) ([Go 1.17 Release Notes - The Go Programming Language](https://go.dev/doc/go1.17#:~:text=Cgo)) ([Go 1.23 Release Notes - The Go Programming Language](https://tip.golang.org/doc/go1.23#:~:text=Cgo)) ([Go 1.16 Release Notes - The Go Programming Language](https://go.dev/doc/go1.16#:~:text=Deprecation%20of%20io%2Fioutil)) ([What's new in Go 1.13](https://blog.kowalczyk.info/article/715712478e8d4dbab6b7c140c8f5e76e/whats-new-in-go-1.13.html#:~:text=To%20create%20an%20error%20that,if%20nothing%20is%20wrapped)) ([Go 1.20 Release Notes - The Go Programming Language](https://tip.golang.org/doc/go1.20#:~:text=The%20new%20function%20errors,wrapping%20a%20list%20of%20errors)) (and others as cited above). These detail the changes in the language, toolchain, and libraries over the versions.


# Conclusion
In conclusion, upgrading the Go Toxcore C library from version 1.12 to 1.23 involves several important changes and improvements. The transition to Go modules, updates in cgo behavior, and enhancements in the standard library all contribute to a more robust and efficient development experience. By following the outlined steps and understanding the changes, developers can successfully upgrade their projects and take advantage of the latest features and optimizations in Go.


---

Please study the attached repomix file of the original go-toxcore-c library to understand the current implementation and how it can be adapted to the new version.

From there please help me to create a plan for upgrading the library to version 1.23, including any necessary code changes, testing strategies, and documentation updates.


Below is a detailed plan outlining how you can rewrite the go‑toxcore‑c library from scratch as a standalone GitHub repository (https://github.com/cammclain/go-toxcore-c), using modern Go (v1.23) best practices. This plan is inspired by the considerations in the attached document citeturn1file0 and covers project setup, design, implementation, testing, and deployment.

---

## 1. Define Requirements & Scope

- **Understand the Tox Protocol:**  
  • Review the Tox specification and core functionalities (client creation, bootstrap, messaging, file transfer, encryption, and callbacks).  
  • Decide which features to support in your initial release versus future iterations.

- **Set Clear Goals:**  
  • Provide a clean, idiomatic, and maintainable Go API.  
  • Eliminate the “slavic garbage” by designing clear, modular, and well-documented code.  
  • Ensure compatibility with Go 1.23, using modules, modern error handling, context, and other language improvements.

---

## 2. Project Setup

- **Repository Initialization:**  
  • Create a new GitHub repository (https://github.com/cammclain/go-toxcore-c).  
  • Initialize the module with:  
    `go mod init github.com/cammclain/go-toxcore-c`  
  • Set the Go version in `go.mod` to `1.23`.

- **Directory Structure:**  
  Organize your project to separate public API, internal cgo bindings, and utility packages. For example:

  ```
  go-toxcore-c/
  ├── go.mod
  ├── README.md
  ├── LICENSE
  ├── cmd/                // Optional: for sample apps and tests
  │   └── example/
  │       └── main.go
  ├── internal/
  │   ├── cgo/            // Low-level cgo bindings to Toxcore C library
  │   │   └── bindings.go
  │   ├── api/            // High-level Go wrapper exposing the public API
  │   │   ├── client.go
  │   │   ├── messages.go
  │   │   ├── files.go    // (if file transfer is in scope)
  │   │   ├── encryption.go
  │   │   └── errors.go
  │   └── utils/          // Helper functions, logging, etc.
  └── pkg/                // Optional: for reusable public packages
      └── toxcore/
          └── toxcore.go  // The main package exposing your library’s API
  ```

- **CI/CD Configuration:**  
  • Set up GitHub Actions (or another CI system) to build, test, and lint your code on every commit.  
  • Ensure your CI uses Go 1.23 and enforces module mode (e.g., using `go mod tidy`, `go test -race`, etc.).

---

## 3. Architecture & API Design

- **Public API Design:**  
  • Define a central type (e.g., `ToxClient`) that encapsulates the Tox client’s state and exposes methods for initialization, bootstrapping, messaging, and shutdown.  
  • Use modern Go error handling (with `%w` for wrapping and, if needed, `errors.Join` for aggregated errors) and context (for cancellation and timeouts).

- **Modularization:**  
  • **Low-Level Bindings:** Create a dedicated package (e.g., `internal/cgo`) that contains all cgo directives and bindings.  
  • **High-Level API:** In a separate package (e.g., `internal/api` or in `pkg/toxcore`), build an idiomatic Go interface that hides the cgo complexity.  
  • **Utilities and Error Handling:** Centralize common functions and custom error definitions to improve readability and maintenance.

- **Best Practices:**  
  • Use Go modules exclusively; avoid legacy GOPATH practices.  
  • Prefer standard library packages (e.g., `os`, `context`, `errors`) and modern idioms.  
  • Consider using generics for any helper functions if they can simplify the code, though they’re optional.

---

## 4. Implementation Strategy

### Phase A: Low-Level cgo Bindings

- **Bindings Implementation:**  
  • Write minimal cgo code to interface with the Toxcore C library.  
  • Follow updated cgo practices:
    - Avoid direct allocation of opaque C structs.
    - Use full struct definitions in the cgo preamble if necessary.
    - Leverage `runtime/cgo.Handle` for safe pointer passing.

- **Build Flags & Cross-compilation:**  
  • Include modern `#cgo` directives for compiler and linker flags (consider `-trimpath` and `-ldflags` improvements from Go 1.23 citeturn1file0).

### Phase B: High-Level API Development

- **Define API Types & Methods:**  
  • Create `ToxClient` with methods for:
    - **Initialization:** Constructing a client using options.
    - **Bootstrap:** Connecting to Tox bootstrap servers.
    - **Messaging:** Sending and receiving text (and later, file) messages.
    - **Event Loop Management:** Methods to run and stop the Tox iteration loop.
    - **Shutdown:** Graceful termination of the client.

- **Error & Context Management:**  
  • Implement error wrapping using `%w`.  
  • Utilize context for long-running operations, cancellations, and timeouts.

### Phase C: Additional Features

- **Encryption & Save Data Handling:**  
  • Implement functions for encrypting/decrypting Tox save data using the new APIs (switch from `io/ioutil` to `os` functions).  
- **Callbacks & Concurrency:**  
  • Design a safe mechanism for registering and handling callbacks from the C layer, ensuring proper concurrency (possibly using channels or sync primitives).

---

## 5. Testing & Quality Assurance

- **Unit Testing:**  
  • Write tests for every package using Go’s `testing` package.  
  • Use table-driven tests to cover edge cases and error conditions.

- **Integration Testing:**  
  • Simulate real-world usage of the Tox client, including bootstrapping and message exchange.  
  • Use a combination of mocks and integration tests that interact with the actual Toxcore library.

- **Fuzz Testing & Race Detection:**  
  • Leverage Go 1.18+ fuzz testing (`go test -fuzz`) for critical functions (especially around parsing and data conversion).  
  • Run tests with the race detector enabled to catch concurrency issues.

- **Static Analysis:**  
  • Integrate linters (e.g., golangci-lint) and use `go vet` to maintain code quality.

---

## 6. Documentation & Examples

- **README & Documentation:**  
  • Write a clear README that explains the project goals, usage, installation, and examples.  
  • Provide godoc comments on all exported functions and types.  
  • Include a CHANGELOG and upgrade notes.

- **Usage Examples:**  
  • Create a simple example in the `cmd/example` directory demonstrating how to initialize a client, bootstrap, send a message, and shut down.

---

## 7. CI/CD & Release Process

- **Automated Testing:**  
  • Configure GitHub Actions to run tests on every commit, pull request, and release candidate.  
  • Ensure the CI builds on multiple operating systems if cross-compilation is a goal.

- **Versioning & Releases:**  
  • Use semantic versioning.  
  • Automate releases (tagging, building, and publishing binaries if applicable).

---

## 8. Iterative Improvement & Community Feedback

- **Alpha/Beta Releases:**  
  • Release early versions to gather feedback.  
  • Iteratively improve the API, fix bugs, and add missing features based on user input.

- **Refactoring:**  
  • Plan periodic refactoring sessions to incorporate new Go features (like generics or additional standard library improvements) as the language evolves.

---

## Conclusion

This plan provides a roadmap for rewriting the go‑toxcore‑c library from scratch with modern Go practices. By starting with a clean project structure, designing a clear public API, implementing safe cgo bindings, and rigorously testing and documenting your code, you can replace the existing codebase with a maintainable, idiomatic solution. This will not only modernize your project but also ensure it is robust and easier to work with in the long run.


Here is the current file tree:

```bash
 $ tree src
 src
└──  go-toxcore-c
    ├──  cmd
    │   └──  example
    │       └──  main.go
    ├──  go.mod
    ├──  internal
    │   ├──  api
    │   │   ├──  client.go
    │   │   ├──  encryption.go
    │   │   ├──  errors.go
    │   │   ├──  files.go
    │   │   └──  messages.go
    │   └──  cgo
    │       └──  bindings.go
    └──  pkg
        └──  toxcore
            └──  toxcore.go
```