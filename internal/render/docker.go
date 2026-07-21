package render

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

// DockerOptions controls human-readable docker output verbosity.
type DockerOptions struct {
	All     bool
	Images  bool
	Volumes bool
}

// WriteDocker renders a human-readable Docker report.
func WriteDocker(w io.Writer, snap *snapshot.Snapshot, noColor bool, opts DockerOptions) error {
	if snap == nil {
		return fmt.Errorf("snapshot is nil")
	}

	label := func(s string) string {
		if noColor {
			return s
		}
		return s
	}

	if _, err := fmt.Fprintf(w, "%s\n", label("bareai docker")); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Collected: %s\n\n", snap.CollectedAt.Format(time.RFC3339)); err != nil {
		return err
	}

	if snap.Docker == nil {
		if _, err := fmt.Fprintln(w, "Docker not available."); err != nil {
			return err
		}
		return writeSkipped(w, snap.Skipped)
	}

	d := snap.Docker
	if !d.Available {
		reason := dockerUnavailableReason(snap.Skipped)
		if _, err := fmt.Fprintf(w, "Docker not available: %s\n", reason); err != nil {
			return err
		}
		return writeSkipped(w, snap.Skipped)
	}

	if err := writeDockerEngine(w, d); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if err := writeDockerContainers(w, d, opts.All); err != nil {
		return err
	}
	if err := writeDockerImages(w, d, opts.Images); err != nil {
		return err
	}
	if err := writeDockerVolumes(w, d, opts.Volumes); err != nil {
		return err
	}

	return writeSkipped(w, snap.Skipped)
}

func dockerUnavailableReason(skips []snapshot.Skip) string {
	for _, skip := range skips {
		if skip.Component == "docker" {
			return skip.Reason
		}
	}
	return "docker daemon not available"
}

func writeDockerEngine(w io.Writer, d *snapshot.Docker) error {
	engine := fmt.Sprintf("Engine: Docker %s  api %s  %s/%s",
		d.ServerVersion, d.APIVersion, d.OSType, d.Architecture)
	if _, err := fmt.Fprintln(w, engine); err != nil {
		return err
	}

	runtime := d.DefaultRuntime
	if runtime == "" {
		runtime = "unknown"
	}
	nvidia := "no"
	if d.NVIDIARuntime {
		nvidia = "yes"
	}
	if _, err := fmt.Fprintf(w, "Runtime: %s (default)  nvidia: %s\n", runtime, nvidia); err != nil {
		return err
	}
	return nil
}

func writeDockerContainers(w io.Writer, d *snapshot.Docker, showAll bool) error {
	running, total := dockerContainerCounts(d.Containers)
	if _, err := fmt.Fprintf(w, "Containers (%d running / %d total)\n", running, total); err != nil {
		return err
	}

	containers := dockerContainersForDisplay(d.Containers, showAll)
	if len(containers) == 0 {
		if _, err := fmt.Fprintln(w, "  none"); err != nil {
			return err
		}
		return nil
	}

	if _, err := fmt.Fprintf(w, "  %-16s %-20s %-10s %-22s %s\n",
		"NAME", "IMAGE", "STATE", "PORTS", "GPU"); err != nil {
		return err
	}

	for _, c := range containers {
		gpu := "no"
		if c.GPURequested {
			gpu = "yes"
		}
		if _, err := fmt.Fprintf(w, "  %-16s %-20s %-10s %-22s %s\n",
			truncate(c.Name, 16),
			truncate(c.Image, 20),
			truncate(c.State, 10),
			truncate(formatDockerPorts(c.Ports), 22),
			gpu,
		); err != nil {
			return err
		}
	}
	return nil
}

func dockerContainerCounts(containers []snapshot.DockerContainer) (running, total int) {
	total = len(containers)
	for _, c := range containers {
		if strings.EqualFold(c.State, "running") {
			running++
		}
	}
	return running, total
}

func dockerContainersForDisplay(containers []snapshot.DockerContainer, showAll bool) []snapshot.DockerContainer {
	if showAll {
		out := append([]snapshot.DockerContainer(nil), containers...)
		sort.Slice(out, func(i, j int) bool {
			return out[i].Name < out[j].Name
		})
		return out
	}

	out := make([]snapshot.DockerContainer, 0, len(containers))
	for _, c := range containers {
		if strings.EqualFold(c.State, "running") {
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Name < out[j].Name
	})
	return out
}

func formatDockerPorts(ports []snapshot.DockerPort) string {
	if len(ports) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(ports))
	for _, p := range ports {
		if p.PublicPort > 0 {
			ip := p.IP
			if ip == "" || ip == "0.0.0.0" || ip == "::" {
				parts = append(parts, fmt.Sprintf("%d->%d/%s", p.PublicPort, p.PrivatePort, p.Type))
			} else {
				parts = append(parts, fmt.Sprintf("%s:%d->%d/%s", ip, p.PublicPort, p.PrivatePort, p.Type))
			}
			continue
		}
		parts = append(parts, fmt.Sprintf("%d/%s", p.PrivatePort, p.Type))
	}
	return strings.Join(parts, ", ")
}

func writeDockerImages(w io.Writer, d *snapshot.Docker, showDetail bool) error {
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if !showDetail {
		if _, err := fmt.Fprintf(w, "Images: %d  (pass --images for detail)\n", len(d.Images)); err != nil {
			return err
		}
		return nil
	}

	if _, err := fmt.Fprintf(w, "Images (%d)\n", len(d.Images)); err != nil {
		return err
	}
	if len(d.Images) == 0 {
		if _, err := fmt.Fprintln(w, "  none"); err != nil {
			return err
		}
		return nil
	}

	if _, err := fmt.Fprintf(w, "  %-14s %-28s %s\n", "ID", "TAGS", "SIZE"); err != nil {
		return err
	}
	for _, img := range d.Images {
		tags := strings.Join(img.RepoTags, ", ")
		if tags == "" {
			tags = "<none>"
		}
		if _, err := fmt.Fprintf(w, "  %-14s %-28s %s\n",
			truncate(img.ID, 14),
			truncate(tags, 28),
			formatBytes(img.Size),
		); err != nil {
			return err
		}
	}
	return nil
}

func writeDockerVolumes(w io.Writer, d *snapshot.Docker, showDetail bool) error {
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if !showDetail {
		if _, err := fmt.Fprintf(w, "Volumes: %d  (pass --volumes for detail)\n", len(d.Volumes)); err != nil {
			return err
		}
		return nil
	}

	if _, err := fmt.Fprintf(w, "Volumes (%d)\n", len(d.Volumes)); err != nil {
		return err
	}
	if len(d.Volumes) == 0 {
		if _, err := fmt.Fprintln(w, "  none"); err != nil {
			return err
		}
		return nil
	}

	if _, err := fmt.Fprintf(w, "  %-24s %-12s %s\n", "NAME", "DRIVER", "MOUNTPOINT"); err != nil {
		return err
	}
	for _, vol := range d.Volumes {
		if _, err := fmt.Fprintf(w, "  %-24s %-12s %s\n",
			truncate(vol.Name, 24),
			truncate(vol.Driver, 12),
			vol.Mountpoint,
		); err != nil {
			return err
		}
	}
	return nil
}

func writeSkipped(w io.Writer, skips []snapshot.Skip) error {
	if len(skips) == 0 {
		return nil
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "Skipped:"); err != nil {
		return err
	}
	for _, skip := range skips {
		if _, err := fmt.Fprintf(w, "  - %s: %s\n", skip.Component, skip.Reason); err != nil {
			return err
		}
	}
	return nil
}

func truncate(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

func writeDockerSummary(w io.Writer, d *snapshot.Docker) error {
	if d == nil || !d.Available {
		if _, err := fmt.Fprintln(w, "Docker:      not available"); err != nil {
			return err
		}
		return nil
	}

	running, _ := dockerContainerCounts(d.Containers)
	line := fmt.Sprintf("Docker:      available — %d running, %d images", running, len(d.Images))
	if d.NVIDIARuntime {
		line += ", nvidia runtime"
	}
	if _, err := fmt.Fprintln(w, line); err != nil {
		return err
	}
	return nil
}
