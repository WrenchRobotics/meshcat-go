package main

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// main refreshes vendored MeshCat viewer assets by building upstream dist
// output at a specific git ref and copying that result into viewer_assets/dist.
func main() {
	var (
		ref     string
		repoURL string
		outDir  string
	)

	flag.StringVar(&ref, "ref", "master", "upstream git ref: branch, tag, or commit hash")
	flag.StringVar(&repoURL, "repo-url", "https://github.com/meshcat-dev/meshcat.git", "upstream MeshCat git repository URL")
	flag.StringVar(&outDir, "out", "viewer_assets/dist", "destination directory for vendored dist files")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: go run ./scripts/refresh_viewer_assets -ref <branch|tag|commit> [options]\n\n")
		fmt.Fprintf(os.Stderr, "Refreshes viewer_assets/dist from an upstream MeshCat build.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	// Validate required external tooling early so failures are immediate and clear.
	for _, tool := range []string{"git", "npm"} {
		if _, err := exec.LookPath(tool); err != nil {
			exitf("required tool not found in PATH: %s", tool)
		}
	}

	absOutDir, err := filepath.Abs(outDir)
	if err != nil {
		exitf("resolve output directory: %v", err)
	}

	// Build in a temp clone to avoid mutating any local checkout of upstream.
	tmpDir, err := os.MkdirTemp("", "meshcat-refresh-*")
	if err != nil {
		exitf("create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	upstreamDir := filepath.Join(tmpDir, "meshcat")

	fmt.Println("Cloning upstream MeshCat repository...")
	runCmd("", "git", "clone", repoURL, upstreamDir)

	fmt.Printf("Checking out ref: %s\n", ref)
	runCmd(upstreamDir, "git", "checkout", ref)

	fmt.Println("Installing npm dependencies...")
	runCmd(upstreamDir, "npm", "install")

	fmt.Println("Building viewer dist assets...")
	runCmd(upstreamDir, "npm", "run", "build")

	// Copy only the generated dist output into the vendored asset location.
	srcDist := filepath.Join(upstreamDir, "dist")
	if st, err := os.Stat(srcDist); err != nil || !st.IsDir() {
		exitf("upstream build did not produce dist directory")
	}

	if err := os.MkdirAll(absOutDir, 0o755); err != nil {
		exitf("create output directory: %v", err)
	}

	if err := clearDir(absOutDir); err != nil {
		exitf("clear output directory: %v", err)
	}

	if err := copyDirContents(srcDist, absOutDir); err != nil {
		exitf("copy built assets: %v", err)
	}

	// Report resolved commit so updates are traceable in commit messages/PRs.
	commit := strings.TrimSpace(runCmdOutput(upstreamDir, "git", "rev-parse", "HEAD"))
	fmt.Printf("Refreshed vendored viewer assets in %s\n", absOutDir)
	fmt.Printf("Upstream commit: %s\n", commit)
}

// runCmd executes a command with stdout/stderr passthrough and exits on error.
func runCmd(dir, name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		exitf("command failed: %s %s (%v)", name, strings.Join(args, " "), err)
	}
}

// runCmdOutput executes a command and returns stdout, exiting on failure.
func runCmdOutput(dir, name string, args ...string) string {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		exitf("command failed: %s %s (%v)", name, strings.Join(args, " "), err)
	}
	return string(out)
}

// clearDir removes all entries within path while preserving the directory itself.
func clearDir(path string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := os.RemoveAll(filepath.Join(path, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}

// copyDirContents recursively copies srcDir contents into dstDir, preserving
// relative structure and file modes.
func copyDirContents(srcDir, dstDir string) error {
	return filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}

		dstPath := filepath.Join(dstDir, rel)
		if d.IsDir() {
			return os.MkdirAll(dstPath, 0o755)
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		info, err := srcFile.Stat()
		if err != nil {
			return err
		}

		dstFile, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
		if err != nil {
			return err
		}
		defer dstFile.Close()

		_, err = io.Copy(dstFile, srcFile)
		return err
	})
}

// exitf prints a formatted error and exits the program with a non-zero status.
func exitf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}
