# 4. Basic architecture follows the Model View Controller pattern

Date: 2025-06-16

## Status

Accepted

## Context

We will follow the Model View Controller pattern for our overall architecture.

## Decision

The model is represented by the Buffer interface in the app module, the view is represented by
the View interface in the app module, and the controller is represented by the App interface in the app module.

## Consequences

Separating the editor into model, view and controller components keeps the code
organized and testable. Each layer can evolve independently, though more
interfaces must be maintained.
