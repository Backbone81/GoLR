# Technology

## Problem statement

To build a modular and extensible LR parser generator, we must choose an implementation language.

The criteria we want to apply for our selection are the following:

- The programming language should be easy to learn and maintain. This is important for a big community of seasoned
  developers as well as new joiners to be able to contribute.
- The programming language should produce fast machine code. This is important to meet the time constraints for
  interactive use.
- The produced binary should rely on as few dependencies during runtime as possible. This is important to make it easy
  to use in as many situations as possible.
- There should be strong support for all major operating systems like Windows, Linux and Mac as well as for all major
  architectures like AMD64 and ARM64.

## Solution

While there are a lot of programming languages out there which meet our requirements to different extends, we decide
to implement GoLR in Go. The reasons are as follows:

- Go is known to be easy to learn and fosters a large community.
- Go's garbage collector frees developers from manual management and reduces the risks of memory related
  bugs. But this advantage also comes with the downside of loosing fine-grained control over the lifetime of objects
  which requires special attention in performance critical code.
- Go source code is compiled into machine code which allows for great execution speed compared to interpreted
  programming languages.
- Go binaries are statically linked and do not need to have any other dependencies during runtime.
- Go supports a wide range of operating systems and CPU architectures.

Other programming languages like C, C++, Rust, Python, Java have been considered and discarded. Some of those
alternatives have a steep learning curve, require manual memory management, require additional runtime dependencies
for execution or are interpreted languages. Go seems to hit the sweet spot of simplicity and performance.
