// Separate module on purpose: the seed tool is an external black-box client.
// It shares no code with the server and talks only to the public HTTP API, so
// it has its own module boundary and depends on nothing but the standard library.
module pixeltracker-seed

go 1.26.3
