# 3. Use Tcell

Date: 2025-06-16

## Status

Accepted

## Context

What package should we use for terminal interaction?

## Decision

We will use the Tcell package, found at https://github.com/gdamore/tcell

## Consequences

Tcell gives the editor rich, portable terminal support. Using it simplifies
handling keyboard input and screen drawing across platforms but introduces an
external dependency that must be kept up to date.

