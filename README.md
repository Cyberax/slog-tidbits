# slog-tidbits

This is a package with extensions for the `slog` package in the standard library. The `slog` package has 
some great ideas, but it lacks a few convenience features that make life so much easier.

# Features

The main feature is debug mode that makes life easier when developing. It allows you to write log messages
in text format, with user-friendly treatment of stack traces (they are multiline). The production mode, on the other
hand, formats stack traces as structured JSON objects.



# Notes

The `slog-tidbits` package is NOT optimized for speed, it's mostly at the proof-of-concept stage right now. So
expect it to be an order of magnitude slower than `slog`. That being said, it still can easily push megabytes of 
log messages per second, which should be more than enough for most applications.

Ideally, some of the `slog-tidbits` features should be integrated into the `slog` package :(

# TODO

1. Rate-limiting for messages, rules for sampling/discarding, and statistics for dropped messages. 
