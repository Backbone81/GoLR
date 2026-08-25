# Parser Generator Backend: Null

This backend writes nothing. See [parser generator backends](parsergen-backend.md) for the other backends.

What it is good for is running the frontend and the core without paying for the output: checking in a build that a
grammar still has no conflicts, and benchmarking the frontend and the core without a backend in the measurement.
