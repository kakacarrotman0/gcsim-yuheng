package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestFullRegenGateOpensOnlyForExplicitOptIn(t *testing.T) {
	t.Setenv("GCSIM_ALLOW_FULL_REGEN", "")
	if err := fullRegenGate(); err == nil || !strings.Contains(err.Error(), "refusing") {
		t.Fatalf("closed gate: %v", err)
	}
	t.Setenv("GCSIM_ALLOW_FULL_REGEN", "1")
	if err := fullRegenGate(); err != nil {
		t.Fatal(err)
	}
}

func TestRunRefusesBeforeCompile(t *testing.T) {
	t.Setenv("GCSIM_ALLOW_FULL_REGEN", "")
	err := run(context.Background(), nil)
	if err == nil || !strings.Contains(err.Error(), "refusing full pipeline regeneration") {
		t.Fatal(err)
	}
	if !strings.Contains(err.Error(), "were not modified") {
		t.Fatal(err)
	}
}

func TestPipelineRunRefusesWithoutTouchingGeneratedFiles(t *testing.T) {
	root := repoRoot(t)
	if err := os.MkdirAll(filepath.Join(root, "pipeline", "_cache"), 0o755); err != nil {
		t.Fatal(err)
	}
	before := gitPorcelain(t, root)

	cmd := exec.Command("go", "run", "./pipeline", "run")
	cmd.Dir = root
	cmd.Env = envWithout("GCSIM_ALLOW_FULL_REGEN")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("pipeline run succeeded:\n%s", out)
	}
	if !bytes.Contains(out, []byte("refusing full pipeline regeneration")) {
		t.Fatalf("output:\n%s", out)
	}
	after := gitPorcelain(t, root)
	if before != after {
		t.Fatalf("tracked tree changed\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

func TestPipelineCleanDoesNotDeleteGeneratedFiles(t *testing.T) {
	dir := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(envWithout("GIT_DIR", "GIT_WORK_TREE"),
			"GIT_AUTHOR_NAME=regen",
			"GIT_AUTHOR_EMAIL=regen@example.com",
			"GIT_COMMITTER_NAME=regen",
			"GIT_COMMITTER_EMAIL=regen@example.com",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
	git("init")
	if err := os.WriteFile(filepath.Join(dir, "sample.dm.go"), []byte("package sample\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git("add", "sample.dm.go")
	git("commit", "-m", "sample")

	script := filepath.Join(repoRoot(t), "scripts", "pipeline-clean.sh")
	for _, allow := range []string{"", "1"} {
		cmd := exec.Command("sh", script)
		cmd.Dir = dir
		env := envWithout("GCSIM_ALLOW_FULL_REGEN")
		if allow != "" {
			env = append(env, "GCSIM_ALLOW_FULL_REGEN="+allow)
		}
		cmd.Env = env
		out, err := cmd.CombinedOutput()
		if err == nil {
			t.Fatalf("pipeline-clean succeeded with GCSIM_ALLOW_FULL_REGEN=%q:\n%s", allow, out)
		}
		if !bytes.Contains(out, []byte("refusing pipeline-clean")) {
			t.Fatalf("output:\n%s", out)
		}
		if _, err := os.Stat(filepath.Join(dir, "sample.dm.go")); err != nil {
			t.Fatalf("generated file removed with GCSIM_ALLOW_FULL_REGEN=%q: %v", allow, err)
		}
		st := gitPorcelain(t, dir)
		if strings.TrimSpace(st) != "" {
			t.Fatalf("temp repo changed with GCSIM_ALLOW_FULL_REGEN=%q:\n%s", allow, st)
		}
	}
}

func TestWorkflowDoesNotCleanBeforeValidation(t *testing.T) {
	root := repoRoot(t)
	wf, err := os.ReadFile(filepath.Join(root, ".github", "workflows", "ci-pipeline.yml"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(wf)
	if strings.Contains(text, "pipeline-clean") || strings.Contains(text, "git rm") {
		t.Fatalf("workflow still deletes generated files:\n%s", text)
	}
	refuse := strings.Index(text, "refusing full pipeline regeneration")
	run := strings.Index(text, "task pipeline --")
	if refuse < 0 || run < 0 || refuse > run {
		t.Fatal("workflow must refuse before task pipeline")
	}
	if !strings.Contains(text, "allow_full_regen") || !strings.Contains(text, "exit 1") {
		t.Fatal("workflow is missing the explicit regen gate")
	}
}

func TestTaskPipelineRefusesWithoutChangingEnum(t *testing.T) {
	root := repoRoot(t)
	enumerPath := filepath.Join(root, "pipeline", "kind_enumer.go")
	beforeEnumer, err := os.ReadFile(enumerPath)
	if err != nil {
		t.Fatal(err)
	}
	before := gitPorcelain(t, root)

	task := taskBin(t)
	cmd := exec.Command(task, "pipeline")
	cmd.Dir = root
	cmd.Env = envWithout("GCSIM_ALLOW_FULL_REGEN")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("task pipeline succeeded:\n%s", out)
	}
	if !bytes.Contains(out, []byte("refusing full pipeline regeneration")) {
		t.Fatalf("output:\n%s", out)
	}
	afterEnumer, err := os.ReadFile(enumerPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(beforeEnumer, afterEnumer) {
		t.Fatal("pipeline/kind_enumer.go changed")
	}
	after := gitPorcelain(t, root)
	if before != after {
		t.Fatalf("worktree changed\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

func TestPipelineTaskChecksGateBeforeAnyGenerate(t *testing.T) {
	text := taskfile(t)
	section, ok := taskSection(text, "pipeline")
	if !ok {
		t.Fatal("pipeline task missing")
	}
	if strings.Contains(section, "go generate") {
		t.Fatalf("normal pipeline task still generates:\n%s", section)
	}
	preflight := strings.Index(section, "sh scripts/pipeline-preflight.sh")
	run := strings.Index(section, "go run ./pipeline")
	if preflight < 0 || run < 0 || preflight > run {
		t.Fatalf("preflight must precede go run:\n%s", section)
	}
	if strings.Count(text, "go generate ./pipeline") != 1 {
		t.Fatalf("go generate ./pipeline count = %d", strings.Count(text, "go generate ./pipeline"))
	}
	enumer, ok := taskSection(text, "pipeline-enumer")
	if !ok {
		t.Fatal("pipeline-enumer task missing")
	}
	if !strings.Contains(enumer, "Unsupported maintenance") || !strings.Contains(enumer, "go generate ./pipeline") {
		t.Fatalf("enum regeneration is not an explicit maintenance task:\n%s", enumer)
	}
	if !strings.Contains(enumer, "sh scripts/pipeline-preflight.sh") {
		t.Fatal("pipeline-enumer skipped the regeneration gate")
	}
}

func TestTaskBuildSucceedsWithoutRegeneration(t *testing.T) {
	root := repoRoot(t)
	before := gitPorcelain(t, root)
	cmd := exec.Command(taskBin(t), "build")
	cmd.Dir = root
	cmd.Env = envWithout("GCSIM_ALLOW_FULL_REGEN")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("task build: %v\n%s", err, out)
	}
	if gitPorcelain(t, root) != before {
		t.Fatalf("task build changed the worktree:\n%s", gitPorcelain(t, root))
	}
}

func TestBuildTaskCompilesCheckedInData(t *testing.T) {
	root := repoRoot(t)
	b, err := os.ReadFile(filepath.Join(root, "Taskfile.yml"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(b)
	start := strings.Index(text, "\n  build:\n")
	end := strings.Index(text, "\n  build-generated:\n")
	if start < 0 || end < 0 || end < start {
		t.Fatal("build tasks missing")
	}
	section := text[start:end]
	if strings.Contains(section, "pipeline") || strings.Contains(section, "protos") {
		t.Fatalf("supported build still regenerates:\n%s", section)
	}
	if !strings.Contains(section, "go build -o {{.OUT | default \"gcsim\"}} ./cmd/gcsim") {
		t.Fatalf("supported build command missing:\n%s", section)
	}
}

func taskSection(text, name string) (string, bool) {
	lines := strings.Split(text, "\n")
	start := -1
	for i, line := range lines {
		if line == "  "+name+":" {
			start = i
			break
		}
	}
	if start < 0 {
		return "", false
	}
	end := len(lines)
	for i := start + 1; i < len(lines); i++ {
		line := lines[i]
		if strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "   ") && strings.HasSuffix(line, ":") && !strings.HasPrefix(line, "  #") {
			end = i
			break
		}
	}
	return strings.Join(lines[start:end], "\n"), true
}

func taskfile(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(repoRoot(t), "Taskfile.yml"))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func taskBin(t *testing.T) string {
	t.Helper()
	p, err := exec.LookPath("task")
	if err != nil {
		t.Fatal("task is not on PATH")
	}
	return p
}

func repoRoot(t *testing.T) string {
	t.Helper()
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}

func gitPorcelain(t *testing.T, dir string) string {
	t.Helper()
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

func envWithout(keys ...string) []string {
	drop := map[string]bool{}
	for _, k := range keys {
		drop[k] = true
	}
	var out []string
	for _, kv := range os.Environ() {
		name, _, _ := strings.Cut(kv, "=")
		if drop[name] {
			continue
		}
		out = append(out, kv)
	}
	return out
}
