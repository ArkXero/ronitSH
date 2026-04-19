// Package content loads and parses markdown files embedded in the binary.
// Files use YAML frontmatter delimited by "---" lines.
//
// Content is loaded once at startup from an embed.FS passed in by the caller
// (cmd/ronit-sh/main.go holds the //go:embed directive at the module root so
// that the embed path resolves correctly).
package content

import (
	"bytes"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"charm.land/glamour/v2"
	"gopkg.in/yaml.v3"
)

// ProjectMeta holds the YAML frontmatter for a project markdown file.
type ProjectMeta struct {
	Title   string            `yaml:"title"`
	Tagline string            `yaml:"tagline"`
	Status  string            `yaml:"status"` // "shipped", "in-progress", "archived"
	Stack   []string          `yaml:"stack"`
	Links   map[string]string `yaml:"links"`
}

// PostMeta holds the YAML frontmatter for a post markdown file.
type PostMeta struct {
	Title string `yaml:"title"`
	Date  string `yaml:"date"` // YYYY-MM-DD
}

// PageMeta holds frontmatter for single-page content (about, now).
type PageMeta struct {
	Title string `yaml:"title"`
	Date  string `yaml:"date"`
}

// Project is a parsed project file.
type Project struct {
	Meta     ProjectMeta
	Body     string // raw markdown body (after frontmatter)
	Filename string // base filename without extension
}

// Post is a parsed post file.
type Post struct {
	Meta     PostMeta
	Body     string
	Filename string // includes date prefix, e.g. "2026-04-building-termfolio"
}

// Page is a parsed single-page file (about, now).
type Page struct {
	Meta PageMeta
	Body string
}

// Loader provides access to parsed embedded content.
type Loader struct {
	fs       fs.FS
	About    Page
	Now      Page
	Projects []Project // sorted by filename
	Posts    []Post    // sorted by filename descending (newest first)
}

// NewLoader parses all content from the given fs.FS.
// Call once at startup; the result is immutable.
func NewLoader(fsys fs.FS) (*Loader, error) {
	// Strip the leading "content" path component so callers use relative paths.
	sub, err := fs.Sub(fsys, "content")
	if err != nil {
		return nil, fmt.Errorf("content sub-fs: %w", err)
	}

	l := &Loader{fs: sub}

	if l.About, err = loadPage(sub, "about.md"); err != nil {
		return nil, fmt.Errorf("about.md: %w", err)
	}
	if l.Now, err = loadPage(sub, "now.md"); err != nil {
		return nil, fmt.Errorf("now.md: %w", err)
	}
	if l.Projects, err = loadProjects(sub); err != nil {
		return nil, fmt.Errorf("projects: %w", err)
	}
	if l.Posts, err = loadPosts(sub); err != nil {
		return nil, fmt.Errorf("posts: %w", err)
	}

	return l, nil
}

// RenderMarkdown renders a markdown string to a terminal-styled string using
// Glamour v2 with the dark style and the given column width for word wrapping.
func RenderMarkdown(md string, width int) (string, error) {
	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return "", err
	}
	return r.Render(md)
}

// --- helpers ---

func loadPage(fsys fs.FS, path string) (Page, error) {
	data, err := fs.ReadFile(fsys, path)
	if err != nil {
		return Page{}, err
	}
	var meta PageMeta
	body, err := parseFrontmatter(data, &meta)
	if err != nil {
		return Page{}, err
	}
	return Page{Meta: meta, Body: body}, nil
}

func loadProjects(fsys fs.FS) ([]Project, error) {
	entries, err := fs.ReadDir(fsys, "projects")
	if err != nil {
		return nil, err
	}

	var projects []Project
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		data, err := fs.ReadFile(fsys, "projects/"+e.Name())
		if err != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), err)
		}
		var meta ProjectMeta
		body, err := parseFrontmatter(data, &meta)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), err)
		}
		projects = append(projects, Project{
			Meta:     meta,
			Body:     body,
			Filename: strings.TrimSuffix(e.Name(), ".md"),
		})
	}

	// Sort by filename for stable ordering.
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Filename < projects[j].Filename
	})

	return projects, nil
}

func loadPosts(fsys fs.FS) ([]Post, error) {
	entries, err := fs.ReadDir(fsys, "posts")
	if err != nil {
		return nil, err
	}

	var posts []Post
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		data, err := fs.ReadFile(fsys, "posts/"+e.Name())
		if err != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), err)
		}
		var meta PostMeta
		body, err := parseFrontmatter(data, &meta)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), err)
		}
		posts = append(posts, Post{
			Meta:     meta,
			Body:     body,
			Filename: strings.TrimSuffix(e.Name(), ".md"),
		})
	}

	// Sort descending by filename (date-prefixed filenames, newest first).
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Filename > posts[j].Filename
	})

	return posts, nil
}

// parseFrontmatter splits raw file bytes into YAML frontmatter and markdown body.
// The frontmatter is unmarshalled into dst. If there is no frontmatter delimiter,
// the entire content is treated as the body.
func parseFrontmatter(raw []byte, dst any) (body string, err error) {
	raw = bytes.TrimSpace(raw)

	// Must start with "---" to have frontmatter.
	if !bytes.HasPrefix(raw, []byte("---")) {
		return string(raw), nil
	}

	// Skip past the opening "---".
	rest := bytes.TrimPrefix(raw, []byte("---"))
	rest = bytes.TrimLeft(rest, "\r\n")

	// Find the closing "---".
	idx := bytes.Index(rest, []byte("\n---"))
	if idx == -1 {
		// No closing delimiter -- treat everything as body.
		return string(raw), nil
	}

	fmBytes := rest[:idx]
	bodyBytes := bytes.TrimSpace(rest[idx+4:]) // skip "\n---"

	if err := yaml.Unmarshal(fmBytes, dst); err != nil {
		return "", fmt.Errorf("frontmatter YAML: %w", err)
	}

	return string(bodyBytes), nil
}
