package purl_test

import (
	"testing"

	"github.com/git-pkgs/purl"
)

func TestGitPURLs(t *testing.T) {
	tests := []struct {
		host, repository, ref, want string
	}{
		{"codeberg.org", "forgejo/forgejo", "a72d2c07cfca03b55371089de6aa230d8c951fa0", "pkg:git/codeberg.org/forgejo/forgejo@a72d2c07cfca03b55371089de6aa230d8c951fa0"},
		{"gitlab.gnome.org", "GNOME/adwaita-fonts", "1.0", "pkg:git/gitlab.gnome.org/GNOME/adwaita-fonts@1.0"},
		{"Git.Example.org", "Group/Subgroup/Repo.git", "", "pkg:git/Git.Example.org/Group/Subgroup/Repo.git"},
		{"cygwin.com", "cgit/newlib-cygwin", "main", "pkg:git/cygwin.com/cgit/newlib-cygwin@main"},
		{"git.example.org", "repo", "refs/heads/main", "pkg:git/git.example.org/repo@refs%2Fheads%2Fmain"},
		{"git.example.org", "team/repo+tools", "release/1.0+build", "pkg:git/git.example.org/team/repo%2Btools@release%2F1.0%2Bbuild"},
		{"git.example.org", "team/repo(!)", "release-1.0", "pkg:git/git.example.org/team/repo%28%21%29@release-1.0"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			coordinate := tt.host + "/" + tt.repository
			made := purl.MakePURL("GIT", coordinate, tt.ref)
			if made == nil {
				t.Fatal("MakePURL returned nil")
			}
			parsed, err := purl.Parse(tt.want)
			if err != nil {
				t.Fatal(err)
			}
			for _, p := range []*purl.PURL{made, parsed, purl.New("git", tt.host, tt.repository, tt.ref, nil)} {
				if p.Namespace != tt.host || p.Name != tt.repository || p.Version != tt.ref {
					t.Errorf("components = %#v", p.PackageURL)
				}
				if p.String() != tt.want || p.ToString() != tt.want {
					t.Errorf("String() = %q, ToString() = %q, want %q", p.String(), p.ToString(), tt.want)
				}
				if p.FullName() != coordinate {
					t.Errorf("FullName() = %q, want %q", p.FullName(), coordinate)
				}
			}
			if got := purl.MakePURLString("GIT", coordinate, tt.ref); got != tt.want {
				t.Errorf("MakePURLString() = %q, want %q", got, tt.want)
			}
			if got := purl.BuildPURLString("GIT", coordinate, tt.ref, ""); got != tt.want {
				t.Errorf("BuildPURLString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGitQualifiersAndSubpath(t *testing.T) {
	const base = "pkg:git/codeberg.org/forgejo/forgejo"
	const qualifiers = "?repository_url=https:%2F%2Fgit.example.org"
	const subpath = "#options/locale%20readme.md"
	p, err := purl.Parse(base + "@main" + qualifiers + subpath)
	if err != nil {
		t.Fatal(err)
	}
	if p.Subpath != "options/locale readme.md" || p.RepositoryURL() != "https://git.example.org" {
		t.Fatalf("components = %#v", p.PackageURL)
	}
	for _, tt := range []struct {
		p    *purl.PURL
		want string
	}{
		{p, base + "@main" + qualifiers + subpath},
		{p.WithVersion("release/v1"), base + "@release%2Fv1" + qualifiers + subpath},
		{p.WithoutVersion(), base + qualifiers + subpath},
		{p.WithoutVersion().WithQualifier("repository_url", "https://mirror.example.org"), base + "?repository_url=https:%2F%2Fmirror.example.org" + subpath},
	} {
		if got := tt.p.String(); got != tt.want {
			t.Errorf("String() = %q, want %q", got, tt.want)
		}
	}
	if got := purl.BuildPURLString("git", "codeberg.org/forgejo/forgejo", "main", "https://git.example.org"); got != base+"@main"+qualifiers {
		t.Errorf("BuildPURLString() = %q", got)
	}
	if p.Version != "main" || p.RepositoryURL() != "https://git.example.org" || p.Namespace != "codeberg.org" || p.Name != "forgejo/forgejo" {
		t.Errorf("original was mutated: %#v", p.PackageURL)
	}
}

func TestGitRequiresHostAndRepository(t *testing.T) {
	for _, name := range []string{"", "codeberg.org", "/forgejo", "codeberg.org/", "codeberg.org//forgejo", "codeberg.org/forgejo/"} {
		t.Run(name, func(t *testing.T) {
			if got := purl.MakePURL("git", name, "main"); got != nil {
				t.Errorf("MakePURL() = %q, want nil", got.String())
			}
			if got := purl.MakePURLString("git", name, "main"); got != "" {
				t.Errorf("MakePURLString() = %q, want empty", got)
			}
			if got := purl.BuildPURLString("git", name, "main", ""); got != "" {
				t.Errorf("BuildPURLString() = %q, want empty", got)
			}
		})
	}
	for _, input := range []string{"pkg:git/repo", "pkg:git/codeberg.org/", "pkg:git/codeberg.org/repo%ZZ"} {
		if _, err := purl.Parse(input); err == nil {
			t.Errorf("Parse(%q) succeeded, want error", input)
		}
	}
}

func TestGitType(t *testing.T) {
	if !purl.IsKnownType("git") || !purl.IsValidEcosystem("GIT") {
		t.Fatal("git is not recognized")
	}
	cfg := purl.TypeInfo("git")
	if cfg == nil || !cfg.NamespaceRequired() || cfg.DefaultRegistry != nil || purl.DefaultRegistry("git") != "" {
		t.Fatalf("unexpected git configuration: %#v", cfg)
	}
}
