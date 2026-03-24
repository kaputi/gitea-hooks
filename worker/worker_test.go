package worker

import (
	"strings"
	"testing"
)

func TestGenerateWorkDir(t *testing.T) {
	w := &Worker{basePath: "/data/reviews"}
	path := w.generateWorkDir("myrepo", 42)

	if !strings.HasPrefix(path, "/data/reviews/myrepo-42-") {
		t.Errorf("path = %q, want prefix '/data/reviews/myrepo-42-'", path)
	}
}
