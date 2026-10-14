//go:build acceptance

package acceptance

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/rogpeppe/go-internal/testscript"
)

// A privileged script writes system paths, makes accounts and starts
// services, and removes them in its last lines. A script that fails stops
// before those lines. Without the sweep, every later script finds a managed
// host and fails too, and the report hides the one real failure under
// hundreds. The sweep removes only what did not exist when the suite
// started, so it never touches an install that the host had before.

var sweepPaths = []string{
	"/etc/safedep/gryph",
	"/etc/claude-code",
	"/etc/codex",
	"/etc/cursor",
	"/etc/gemini-cli",
	"/etc/devin",
	"/etc/gryph-acceptance-trap.json",
	"/usr/local/bin/gryph-acceptance",
	"/usr/share/polkit-1/actions/io.safedep.gryph.policy",
	"/etc/systemd/system/gryph-supervisor.socket",
	"/etc/systemd/system/gryph-supervisor.service",
	"/etc/systemd/system/gryph-fanotify.service",
	"/etc/systemd/user/gryph-reconcile.service",
	"/etc/systemd/user/gryph-reconcile.timer",
	"/etc/systemd/user/timers.target.wants/gryph-reconcile.timer",
	"/var/lib/safedep/gryph",
	"/var/spool/safedep/gryph",
	"/run/safedep/gryph",
	"/root/.config/safedep",
	"/root/.local/share/safedep",
	"/root/.cache/safedep",
}

// sweepParents go only when they are empty, as the scripts remove them.
var sweepParents = []string{"/etc/safedep", "/var/lib/safedep", "/var/spool/safedep", "/run/safedep"}

var (
	sweepAccounts = []string{"gryphacc1", "gryphacc2", "gryphacc3", "_gryph"}
	sweepGroups   = []string{"gryph-acceptance-admins"}
)

const (
	sweepTempGlob = "/tmp/gryph-acceptance-*"
	sweepPIDFile  = "/tmp/gryph-acceptance-state/supervisor.pid"
)

type sweep struct {
	bin  string
	kept map[string]bool
}

func newSweep(bin string) (*sweep, error) {
	s := &sweep{bin: bin, kept: map[string]bool{}}
	paths, err := sweepTargets()
	if err != nil {
		return nil, err
	}
	for _, p := range append(paths, sweepParents...) {
		if _, err := os.Lstat(p); err == nil {
			s.kept[p] = true
		}
	}
	for _, a := range sweepAccounts {
		if _, err := user.Lookup(a); err == nil {
			s.kept["account:"+a] = true
		}
	}
	for _, g := range sweepGroups {
		if _, err := user.LookupGroup(g); err == nil {
			s.kept["group:"+g] = true
		}
	}
	return s, nil
}

func sweepTargets() ([]string, error) {
	tmp, err := filepath.Glob(sweepTempGlob)
	if err != nil {
		return nil, err
	}
	return append(append([]string{}, sweepPaths...), tmp...), nil
}

// run removes what the script left. testscript calls it before it stops
// the background commands of the script, so it stops the decision service
// of the script first: a running service could write its state again. A
// cleanup that fails fails the script, so the report names the script that
// leaves the host dirty for the next one.
func (s *sweep) run(t testscript.T) {
	var removed, failed []string
	if _, err := os.Stat(sweepPIDFile); err == nil && !s.kept[filepath.Dir(sweepPIDFile)] {
		if err := s.exec(s.bin, "supervisor", "stop", "--pid-file", sweepPIDFile); err != nil {
			failed = append(failed, err.Error())
		}
	}
	for _, a := range sweepAccounts {
		if s.kept["account:"+a] {
			continue
		}
		if _, err := user.Lookup(a); err != nil {
			continue
		}
		if err := s.exec("pkill", "-KILL", "-u", a); err != nil {
			failed = append(failed, err.Error())
		}
		if err := s.exec("userdel", "-r", "-f", a); err != nil {
			failed = append(failed, err.Error())
		}
		if _, err := user.Lookup(a); err == nil {
			failed = append(failed, "account "+a+" still exists")
			continue
		}
		removed = append(removed, "account "+a)
	}
	for _, g := range sweepGroups {
		if s.kept["group:"+g] {
			continue
		}
		if _, err := user.LookupGroup(g); err != nil {
			continue
		}
		if err := s.exec("groupdel", g); err != nil {
			failed = append(failed, err.Error())
			continue
		}
		removed = append(removed, "group "+g)
	}
	paths, err := sweepTargets()
	if err != nil {
		failed = append(failed, err.Error())
	}
	for _, p := range paths {
		if s.kept[p] {
			continue
		}
		if _, err := os.Lstat(p); err != nil {
			continue
		}
		if err := os.RemoveAll(p); err != nil {
			failed = append(failed, err.Error())
			continue
		}
		removed = append(removed, p)
	}
	for _, p := range sweepParents {
		if s.kept[p] {
			continue
		}
		if err := os.Remove(p); err == nil {
			removed = append(removed, p)
		}
	}
	if len(removed) > 0 {
		t.Log("sweep: the script left " + strings.Join(removed, ", "))
	}
	if len(failed) > 0 {
		t.Fatal("sweep: cleanup failed, the next script can find a dirty host: " + strings.Join(failed, "; "))
	}
}

// exec runs one cleanup command. pkill exits 1 when no process matched,
// which is the normal case.
func (s *sweep) exec(name string, args ...string) error {
	out, err := exec.Command(name, args...).CombinedOutput()
	var ee *exec.ExitError
	if err == nil || (name == "pkill" && errors.As(err, &ee) && ee.ExitCode() == 1) {
		return nil
	}
	return fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(string(out)))
}
