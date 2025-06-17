# 5. Use the Rope data structure for buffer contents

Date: 2025-06-16

## Status

Accepted

## Context

Rope is a space and time efficient data structure for long strings, and is idempotent, which is important
for undo/redo and overall simplicity.

## Decision

We will use rope data structures for the underlying character data in the text editor.

## Consequences

Ropes make insertions and deletions efficient even for large files, enabling
undo/redo functionality without excessive memory use. The downside is slightly
more complex code when manipulating the buffer directly.
