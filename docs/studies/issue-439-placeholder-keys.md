# Issue 439: placeholder acceptance keys

The existing `AcceptPlaceholder` inserts the entire hint in one change and returns success after the callback. A regression run before the guard failed for Tab, End and Ctrl+E: none inserted the hint. The guard must precede configured bindings and preserve fallback routing when it does not apply.
