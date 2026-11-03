package updater

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

const (
	Owner          = "espresso20"
	Repo           = "acronomicon"
	Branch         = "main"
	ModulePath     = "github.com/espresso20/acronomicon"
	DefaultAPIBase = "https://api.github.com"
)

type Method string

const (
	MethodGit     Method = "git"
	MethodRelease Method = "release"
	MethodGo      Method = "go"
)

type Status struct {
	Available bool
	Current   string
	Latest    string
	Method    Method
}

type Updater struct {
	APIBase   string
	Client    *http.Client
	Version   string
	ExePath   string
	RepoDir   string
	Out       io.Writer
	CacheFile string
	revision  string // unexported: for testing
}

type release struct {
	TagName string  `json:"tag_name"`
	Assets  []asset `json:"assets"`
}

type asset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
	Size int64  `json:"size"`
}

var errNoRelease = errors.New("no published release")

func New(version string) *Updater {
	if version == "" {
		version = "dev"
	}

	exePath := ""
	if exe, err := os.Executable(); err == nil {
		if resolved, err := filepath.EvalSymlinks(exe); err == nil {
			exePath = resolved
		}
	}

	cacheDir, _ := os.UserCacheDir()
	cacheFile := filepath.Join(cacheDir, "acronomicon", "update-check.json")

	u := &Updater{
		APIBase:   DefaultAPIBase,
		Client:    &http.Client{Timeout: 30 * time.Second},
		Version:   version,
		ExePath:   exePath,
		RepoDir:   findRepo(exePath),
		Out:       os.Stdout,
		CacheFile: cacheFile,
	}
	return u
}

func findRepo(exePath string) string {
	var candidates []string
	if exePath != "" {
		candidates = append(candidates, filepath.Dir(exePath))
	}
	if dir, err := os.Getwd(); err == nil {
		candidates = append(candidates, dir)
	}

	if _, err := exec.LookPath("git"); err != nil {
		return ""
	}

	for _, dir := range candidates {
		cmd := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel")
		out, err := cmd.Output()
		if err != nil {
			continue
		}
		top := strings.TrimSpace(string(out))

		cmd = exec.Command("git", "-C", top, "remote", "get-url", "origin")
		urlOut, err := cmd.Output()
		if err != nil {
			continue
		}
		url := strings.ToLower(strings.TrimSpace(string(urlOut)))
		if strings.Contains(url, "espresso20/acronomicon") {
			return top
		}
	}
	return ""
}

var (
	releaseRe = regexp.MustCompile(`^v?\d+\.\d+\.\d+`)
	// pseudo-versions from `go install ...@latest` end in a 12-char commit hash,
	// e.g. v0.0.0-20261103120000-f70b7f4abc12
	pseudoRe = regexp.MustCompile(`-([0-9a-f]{12})$`)
)

func isReleaseVersion(v string) bool {
	return releaseRe.MatchString(v)
}

func compareVersions(a, b string) int {
	parseVer := func(s string) [3]int {
		s = strings.TrimPrefix(s, "v")
		if idx := strings.IndexAny(s, "-+"); idx >= 0 {
			s = s[:idx]
		}
		parts := strings.Split(s, ".")
		var nums [3]int
		for i := 0; i < 3 && i < len(parts); i++ {
			var n int
			fmt.Sscanf(parts[i], "%d", &n)
			nums[i] = n
		}
		return nums
	}

	aNums := parseVer(a)
	bNums := parseVer(b)

	for i := 0; i < 3; i++ {
		if aNums[i] < bNums[i] {
			return -1
		}
		if aNums[i] > bNums[i] {
			return 1
		}
	}
	return 0
}

func (u *Updater) currentRevision() string {
	if u.revision != "" {
		return u.revision
	}

	if u.RepoDir != "" {
		cmd := exec.Command("git", "-C", u.RepoDir, "rev-parse", "HEAD")
		out, err := cmd.Output()
		if err == nil {
			return strings.TrimSpace(string(out))
		}
	}

	info, ok := debug.ReadBuildInfo()
	if ok && info != nil {
		for _, s := range info.Settings {
			if s.Key == "vcs.revision" {
				return s.Value
			}
		}
		// go install builds carry no vcs settings, only a module pseudo-version
		if m := pseudoRe.FindStringSubmatch(info.Main.Version); m != nil {
			return m[1]
		}
	}

	return ""
}

func short(s string) string {
	if len(s) <= 7 {
		return s
	}
	return s[:7]
}

func (u *Updater) latestRelease(ctx context.Context) (*release, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", u.APIBase, Owner, Repo)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "acronomicon-updater")
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := u.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, errNoRelease
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("github api returned status %d", resp.StatusCode)
	}

	var rel release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, err
	}
	return &rel, nil
}

func (u *Updater) compare(ctx context.Context, base string) (int, string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/compare/%s...%s", u.APIBase, Owner, Repo, base, Branch)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, "", err
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "acronomicon-updater")
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := u.Client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return 0, "", nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, "", fmt.Errorf("github api returned status %d", resp.StatusCode)
	}

	var data struct {
		Status  string `json:"status"`
		AheadBy int    `json:"ahead_by"`
		Commits []struct {
			SHA string `json:"sha"`
		} `json:"commits"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, "", err
	}

	latestSHA := base
	if len(data.Commits) > 0 {
		latestSHA = data.Commits[len(data.Commits)-1].SHA
	}

	return data.AheadBy, latestSHA, nil
}

func assetName(goos, goarch string) []string {
	switch goos {
	case "windows":
		return []string{fmt.Sprintf("acronomicon-windows-%s.exe", goarch)}
	case "darwin":
		return []string{
			fmt.Sprintf("acronomicon-darwin-%s", goarch),
			"acronomicon-macos-universal",
		}
	default:
		return []string{fmt.Sprintf("acronomicon-%s-%s", goos, goarch)}
	}
}

func (u *Updater) method() Method {
	if u.RepoDir != "" {
		return MethodGit
	}
	if _, err := exec.LookPath("go"); err == nil {
		return MethodGo
	}
	return MethodRelease
}

func (u *Updater) Check(ctx context.Context) (Status, error) {
	if isReleaseVersion(u.Version) && u.RepoDir == "" {
		rel, err := u.latestRelease(ctx)
		if err == nil && rel != nil {
			current := u.Version
			latest := rel.TagName
			available := compareVersions(latest, current) > 0
			return Status{
				Available: available,
				Current:   current,
				Latest:    latest,
				Method:    MethodRelease,
			}, nil
		}
		if !errors.Is(err, errNoRelease) {
			return Status{}, err
		}
	}

	rev := u.currentRevision()
	if rev == "" {
		return Status{
			Current: "unknown",
			Method:  u.method(),
		}, nil
	}

	behind, latest, err := u.compare(ctx, rev)
	if err != nil {
		return Status{}, err
	}

	return Status{
		Available: behind > 0,
		Current:   short(rev),
		Latest:    short(latest),
		Method:    u.method(),
	}, nil
}

func (u *Updater) CheckCached(ctx context.Context, ttl time.Duration) (Status, error) {
	if u.CacheFile != "" {
		data, err := os.ReadFile(u.CacheFile)
		if err == nil {
			var cached struct {
				CheckedAt time.Time `json:"checked_at"`
				Current   string    `json:"current"`
				Latest    string    `json:"latest"`
				Available bool      `json:"available"`
			}
			if err := json.Unmarshal(data, &cached); err == nil {
				if time.Since(cached.CheckedAt) < ttl {
					currentIdentity := u.Version
					if !isReleaseVersion(u.Version) {
						currentIdentity = short(u.currentRevision())
					}
					if cached.Current == currentIdentity {
						return Status{
							Available: cached.Available,
							Current:   cached.Current,
							Latest:    cached.Latest,
							Method:    u.method(),
						}, nil
					}
				}
			}
		}
	}

	st, err := u.Check(ctx)
	if err != nil {
		return st, err
	}

	if u.CacheFile != "" {
		cacheDir := filepath.Dir(u.CacheFile)
		os.MkdirAll(cacheDir, 0o755)

		data := map[string]interface{}{
			"checked_at": time.Now(),
			"current":    st.Current,
			"latest":     st.Latest,
			"available":  st.Available,
		}
		if bytes, err := json.Marshal(data); err == nil {
			os.WriteFile(u.CacheFile, bytes, 0o644)
		}
	}

	return st, nil
}

func (u *Updater) Update(ctx context.Context) error {
	cleanupOld(u.ExePath)

	if isReleaseVersion(u.Version) && u.RepoDir == "" {
		err := u.updateRelease(ctx)
		if err == nil {
			return nil
		}
		if _, err := exec.LookPath("go"); err != nil {
			return fmt.Errorf("release download failed: %w; install Go or download manually from https://github.com/espresso20/acronomicon", err)
		}
		fmt.Fprintln(u.Out, err)
	}

	switch u.method() {
	case MethodGit:
		return u.updateGit(ctx)
	case MethodGo:
		st, err := u.Check(ctx)
		if err == nil && !st.Available && st.Current != "unknown" {
			fmt.Fprintln(u.Out, "Already up to date.")
			return nil
		}
		return u.updateGo(ctx)
	default:
		if err := u.updateRelease(ctx); err != nil {
			return fmt.Errorf("release download failed: %w; install Go or download manually from https://github.com/espresso20/acronomicon", err)
		}
		return nil
	}
}

func (u *Updater) updateGit(ctx context.Context) error {
	run := func(name string, args ...string) error {
		cmd := exec.CommandContext(ctx, name, args...)
		cmd.Dir = u.RepoDir
		cmd.Stdout = u.Out
		cmd.Stderr = u.Out
		return cmd.Run()
	}

	output := func(name string, args ...string) (string, error) {
		cmd := exec.CommandContext(ctx, name, args...)
		cmd.Dir = u.RepoDir
		out, err := cmd.Output()
		return strings.TrimSpace(string(out)), err
	}

	status, _ := output("git", "status", "--porcelain", "--untracked-files=no")
	if status != "" {
		return fmt.Errorf("local changes in %s; commit or stash them before updating", u.RepoDir)
	}

	fmt.Fprintln(u.Out, "Fetching origin...")
	if err := run("git", "fetch", "--quiet", "origin"); err != nil {
		return err
	}

	old, _ := output("git", "rev-parse", "HEAD")
	upstream, err := output("git", "rev-parse", "@{u}")
	if err != nil {
		upstream, _ = output("git", "rev-parse", "origin/main")
	}

	if old == upstream {
		fmt.Fprintln(u.Out, "Already up to date.")
		return nil
	}

	if err := run("git", "pull", "--ff-only"); err != nil {
		return fmt.Errorf("branch has diverged from origin; resolve manually")
	}

	if _, err := exec.LookPath("go"); err == nil {
		fmt.Fprintln(u.Out, "Building...")
		exeSuffix := ""
		if runtime.GOOS == "windows" {
			exeSuffix = ".exe"
		}

		var target string
		if u.ExePath != "" {
			if rel, err := filepath.Rel(u.RepoDir, u.ExePath); err == nil && !strings.HasPrefix(rel, "..") {
				target = u.ExePath
			}
		}
		if target == "" {
			target = filepath.Join(u.RepoDir, "acronomicon"+exeSuffix)
		}

		tmp := target + ".new"
		cmd := exec.CommandContext(ctx, "go", "build", "-o", tmp, ".")
		cmd.Dir = u.RepoDir
		cmd.Stdout = u.Out
		cmd.Stderr = u.Out
		if err := cmd.Run(); err != nil {
			return err
		}

		if err := swapBinary(tmp, target); err != nil {
			return err
		}

		fmt.Fprintf(u.Out, "Updated acronomicon at %s. Restart to use the new version.\n", target)
		return nil
	}

	fmt.Fprintln(u.Out, "Go not found; source updated but binary not rebuilt. Run: go build -o acronomicon .")
	return nil
}

func (u *Updater) updateRelease(ctx context.Context) error {
	rel, err := u.latestRelease(ctx)
	if err != nil {
		return err
	}

	if isReleaseVersion(u.Version) && compareVersions(rel.TagName, u.Version) <= 0 {
		fmt.Fprintln(u.Out, "Already up to date.")
		return nil
	}

	if u.ExePath == "" {
		return fmt.Errorf("cannot update: executable path not found")
	}

	names := assetName(runtime.GOOS, runtime.GOARCH)
	var downloadAsset *asset
	for _, name := range names {
		for i := range rel.Assets {
			if rel.Assets[i].Name == name {
				downloadAsset = &rel.Assets[i]
				break
			}
		}
		if downloadAsset != nil {
			break
		}
	}

	if downloadAsset == nil {
		return fmt.Errorf("no release asset for %s/%s in %s", runtime.GOOS, runtime.GOARCH, rel.TagName)
	}

	fmt.Fprintf(u.Out, "Downloading %s (%s)...\n", downloadAsset.Name, rel.TagName)

	tmp := u.ExePath + ".new"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", downloadAsset.URL, nil)
	if err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	req.Header.Set("User-Agent", "acronomicon-updater")

	resp, err := u.Client.Do(req)
	if err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		f.Close()
		os.Remove(tmp)
		return fmt.Errorf("failed to download: status %d", resp.StatusCode)
	}

	written, err := io.Copy(f, resp.Body)
	f.Close()

	if err != nil {
		os.Remove(tmp)
		return err
	}

	if downloadAsset.Size > 0 && written != downloadAsset.Size {
		os.Remove(tmp)
		return fmt.Errorf("incomplete download")
	}

	if runtime.GOOS != "windows" {
		os.Chmod(tmp, 0o755)
	}

	if err := swapBinary(tmp, u.ExePath); err != nil {
		os.Remove(tmp)
		return err
	}

	fmt.Fprintf(u.Out, "Updated acronomicon at %s. Restart to use the new version.\n", u.ExePath)
	return nil
}

func (u *Updater) updateGo(ctx context.Context) error {
	if u.ExePath == "" {
		return fmt.Errorf("cannot update: executable path not found")
	}

	fmt.Fprintln(u.Out, "Building...")

	dir, err := os.MkdirTemp("", "acronomicon-update-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	env := os.Environ()
	env = append(env, "GOBIN="+dir, "GOFLAGS=-mod=mod")

	exeSuffix := ""
	if runtime.GOOS == "windows" {
		exeSuffix = ".exe"
	}

	cmd := exec.CommandContext(ctx, "go", "install", ModulePath+"@latest")
	cmd.Env = env
	cmd.Dir = dir
	cmd.Stdout = u.Out
	cmd.Stderr = u.Out

	if err := cmd.Run(); err != nil {
		return err
	}

	built := filepath.Join(dir, "acronomicon"+exeSuffix)
	tmp := u.ExePath + ".new"

	src, err := os.Open(built)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	_, err = io.Copy(dst, src)
	if cerr := dst.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		os.Remove(tmp)
		return err
	}

	if err := swapBinary(tmp, u.ExePath); err != nil {
		return err
	}

	fmt.Fprintf(u.Out, "Updated acronomicon at %s. Restart to use the new version.\n", u.ExePath)
	return nil
}
