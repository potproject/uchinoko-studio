package prompts

import (
	"embed"
	"fmt"
	"strings"
)

//go:embed memory/*.txt
var files embed.FS

var (
	MemoryExtractTurnSystem    = mustRead("memory/extract_turn_system.txt")
	MemoryCompactSummarySystem = mustRead("memory/compact_summary_system.txt")
	MemoryCompactExtractSystem = mustRead("memory/compact_extract_system.txt")
	MemoryPolicy               = mustRead("memory/policy.txt")
)

func mustRead(name string) string {
	body, err := files.ReadFile(name)
	if err != nil {
		panic(fmt.Errorf("read embedded prompt %q: %w", name, err))
	}
	return strings.TrimSpace(string(body))
}
