---
description: Review Go code and suggest improvements
allowed-tools: Read, Grep, Glob
argument-hint: <file-path>
---

# Go Code Review

## Target File

$ARGUMENTS

## Review Process

1. Read the target file
2. Review based on go-code-review skill
3. Present issues and improvements from the following perspectives

## Review Perspectives

### Required Check Items

- [ ] Are errors handled appropriately?
- [ ] Are resources (files, connections) properly closed?
- [ ] Is concurrency safe? (data races, goroutine leaks)
- [ ] Is input validation performed?

### Code Style

- [ ] Are names clear?
- [ ] Are functions appropriately sized?
- [ ] Is early return used?
- [ ] Are comments appropriate?

### Performance

- [ ] Are there unnecessary memory allocations?
- [ ] Are large structs passed by pointer?

### Security

- [ ] SQL injection prevention
- [ ] Check for logging sensitive information

## Output Format

````markdown
## Review Results: [File Name]

### Good Points

- ...

### Points Needing Improvement

#### 1. [Issue Summary] (Priority: High/Medium/Low)

**Location**: Line XX
**Issue**: ...
**Improvement**:

```go
// Improved code
```
````

### Overall Comments

...

```

```
