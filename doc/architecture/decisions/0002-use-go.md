# 2. Use Go

Date: 2025-06-16

## Status

Accepted

## Context

What language should we use and what basic guidelines should we follow?

## Decision

We will use the Go programming language.

We will write unit tests, using the standard Go naming convention and putting
those tests in the corresponding name_test.go files.

We will follow a standard directory structure for Go programs, specifically:

tked/
|--- go.mod
|--- go.sum
|--- LICENSE
|--- README.md
|--- cmd/
|    └── tked/			# this directory is for the main executable for the tked program
|        └── main.go
|--- internal/			# this directory is for packages that are internal to the tked project
|    └── app/			# the app package implements the main model, view, and controller for the tked project
|    └── rope/			# the rope package implements a rope data structure
|        └── rope.go
|        └── rope_test.go
|--- pkg/			# if this project creates any external, reusable packages they will go here
|--- doc/			# documentation (in markdown) for the tked project
|    └── architecture/		# architecture documentation
|        └── decisions/		# architecture decision documents (ADR) live here


## Consequences

Selecting Go gives us a simple toolchain and a thriving ecosystem. The language is
portable across platforms and encourages clear, testable code. Contributors will
need basic familiarity with Go tooling such as `go build` and `go test`.

