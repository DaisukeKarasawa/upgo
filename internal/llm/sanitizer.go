package llm

import (
	"regexp"
	"strings"
)

// sanitizeInput sanitizes user input to prevent prompt injection attacks.
// It escapes formatting tokens, removes suspicious LLM instruction patterns,
// and wraps content in a safe format.
func sanitizeInput(input string) string {
	if input == "" {
		return input
	}

	// Step 1: Escape formatting tokens that could be used in template systems
	// Note: % escaping is not needed when sanitized content is used as %s arguments in fmt.Sprintf
	// Escape curly braces that might be used in template systems
	input = strings.ReplaceAll(input, "{", "\\{")
	input = strings.ReplaceAll(input, "}", "\\}")

	// Step 2: Escape backticks and triple dashes that could be used to create code blocks
	input = strings.ReplaceAll(input, "```", "\\`\\`\\`")
	input = strings.ReplaceAll(input, "---", "\\-\\-\\-")

	// Step 3: Remove or neutralize obvious LLM instruction patterns
	// These patterns are common in prompt injection attacks
	input = removeInstructionPatterns(input)

	// Step 4: Wrap content in a safe format (fenced code block)
	// This clearly separates user content from instructions
	return wrapInCodeBlock(input)
}

// removeInstructionPatterns removes or neutralizes suspicious LLM instruction patterns.
func removeInstructionPatterns(input string) string {
	lines := strings.Split(input, "\n")
	var sanitizedLines []string

	// Patterns that indicate potential prompt injection attempts
	instructionPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)^\s*(ignore|forget|disregard|skip)\s+(the|all|previous|above|following|next)`),
		regexp.MustCompile(`(?i)^\s*(do\s+not|don't|never)\s+(follow|execute|run|use|apply)`),
		regexp.MustCompile(`(?i)^\s*(follow|execute|run|use|apply)\s+(the\s+)?(next|following|new|different)`),
		regexp.MustCompile(`(?i)^\s*(system|assistant|user):\s*`),
		regexp.MustCompile(`(?i)^\s*(you\s+are|you're|act\s+as|pretend\s+to\s+be)`),
		regexp.MustCompile(`(?i)^\s*(output|respond|reply|answer)\s+(only|just|exactly|precisely)`),
		regexp.MustCompile(`(?i)^\s*(delete|remove|clear|erase)\s+(all|everything|previous|above)`),
		regexp.MustCompile(`(?i)^\s*(override|replace|change)\s+(the\s+)?(prompt|instruction|system)`),
	}

	for _, line := range lines {
		shouldRemove := false
		for _, pattern := range instructionPatterns {
			if pattern.MatchString(line) {
				shouldRemove = true
				break
			}
		}

		if !shouldRemove {
			sanitizedLines = append(sanitizedLines, line)
		}
	}

	return strings.Join(sanitizedLines, "\n")
}

// wrapInCodeBlock wraps content in a fenced code block to clearly separate
// user content from LLM instructions. This makes it harder for injected
// instructions to be interpreted as part of the prompt.
func wrapInCodeBlock(content string) string {
	if content == "" {
		return content
	}

	// Use a code block with a specific language marker to indicate this is user content
	return "```user_content\n" + content + "\n```"
}
