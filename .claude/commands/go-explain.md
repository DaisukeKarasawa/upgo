---
description: Explain Go code behavior and design intent
allowed-tools: Read, Grep, Glob
argument-hint: <file-path or code-snippet>
---

# Go Code Explain

## Target

$ARGUMENTS

## Explanation Process

1. Read the target code
2. Explain based on Go's design philosophy
3. Explain from the following perspectives

## Explanation Perspectives

### 1. Overview

- What does this code do?
- Why is it implemented this way? (Relationship with Go's philosophy)

### 2. Detailed Explanation

- Role and behavior of each part
- Patterns and idioms used
- Important Go features (interface, goroutine, channel, etc.)

### 3. Design Points

- Why was this design chosen?
- Comparison with alternatives
- Trade-offs

### 4. Learning Points

- Go best practices that can be learned from this code
- Related Go Proverbs
- Skills to reference

## Output Format

````markdown
## Code Explanation: [Target]

### Overview

This code...

### Detailed Explanation

#### [Section 1]

```go
// Relevant code
```
````

This part...

#### [Section 2]

...

### Design Points

- **Why this design**: ...
- **Relationship with Go's philosophy**: ...

### Learning Points

1. **[Point 1]**: ...
2. **[Point 2]**: ...

### Related Skills

- go-philosophy: ...
- go-error-handling: ...

```

```
