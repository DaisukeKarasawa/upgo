---
name: go-gerrit-reference
description: |
  Reference knowledge for Gerrit API usage with golang/go repository.
  Provides helper functions, authentication setup, common queries, and error handling patterns.
  Use when working with Gerrit API, setting up authentication, or handling Gerrit API errors.
user-invocable: false
---

# Gerrit API Reference

Shared reference for Gerrit REST API usage with golang/go repository. This skill provides common patterns, helper functions, and error handling guidance.

## Quick Start

For Gerrit API helper function and authentication setup, see [REFERENCE.md](REFERENCE.md).

## What This Provides

- **Helper Function**: `gerrit_api()` function that handles XSSI prefix stripping and authentication
- **Authentication Setup**: Environment variables and credential configuration
- **Common Queries**: Typical Gerrit API query patterns for fetching changes
- **Error Handling**: Common error scenarios and how to handle them

## When to Use

This is a reference skill that other skills (`go-pr-fetcher`, `go-pr-analyzer`) and commands reference for Gerrit API patterns. You don't need to invoke this skill directly—it's automatically available as background knowledge.

## Related Skills

- `go-pr-fetcher`: Uses this reference for fetching Change data
- `go-pr-analyzer`: Uses this reference for fetching Change details during analysis

For complete API reference and examples, see [REFERENCE.md](REFERENCE.md).
