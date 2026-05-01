package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"strings"
	"time"
)

var version = "dev"

const (
	boldBlue  = "\033[1;34m"
	lightBlue = "\033[0;94m"
	dim       = "\033[2m"
	yellow    = "\033[0;33m"
	reset     = "\033[0m"
)

type statusInput struct {
	Workspace struct {
		CurrentDir string `json:"current_dir"`
	} `json:"workspace"`
	Model struct {
		DisplayName string `json:"display_name"`
		ID          string `json:"id"`
	} `json:"model"`
	ContextWindow struct {
		UsedPercentage    float64 `json:"used_percentage"`
		ContextWindowSize int64   `json:"context_window_size"`
		TotalInputTokens  int64   `json:"total_input_tokens"`
	} `json:"context_window"`
	RateLimits struct {
		FiveHour struct {
			UsedPercentage float64 `json:"used_percentage"`
			ResetsAt       int64   `json:"resets_at"`
		} `json:"five_hour"`
		SevenDay struct {
			UsedPercentage float64 `json:"used_percentage"`
			ResetsAt       int64   `json:"resets_at"`
		} `json:"seven_day"`
	} `json:"rate_limits"`
	Cost struct {
		TotalCostUSD float64 `json:"total_cost_usd"`
	} `json:"cost"`
	Cwd string `json:"cwd"`
}

func shortenPath(p string) string {
	home, _ := os.UserHomeDir()
	if home != "" && strings.HasPrefix(p, home) {
		p = "~" + p[len(home):]
	}
	if strings.Count(p, "/") > 2 {
		parts := strings.Split(p, "/")
		var segs []string
		for _, s := range parts {
			if s != "" {
				segs = append(segs, s)
			}
		}
		if len(segs) >= 2 {
			return ".../" + segs[len(segs)-2] + "/" + segs[len(segs)-1]
		}
	}
	return p
}

func gitInfo(cwd string) string {
	check := exec.Command("git", "-C", cwd, "rev-parse", "--is-inside-work-tree")
	check.Stderr = io.Discard
	if check.Run() != nil {
		return ""
	}

	var branch string
	if out, err := exec.Command("git", "-C", cwd, "symbolic-ref", "--short", "HEAD").Output(); err == nil {
		branch = strings.TrimSpace(string(out))
	} else if out, err := exec.Command("git", "-C", cwd, "rev-parse", "--short", "HEAD").Output(); err == nil {
		branch = strings.TrimSpace(string(out))
	}
	if branch == "" {
		return ""
	}

	dirty := ""
	d1 := exec.Command("git", "-C", cwd, "diff", "--quiet")
	d1.Stderr = io.Discard
	d2 := exec.Command("git", "-C", cwd, "diff", "--cached", "--quiet")
	d2.Stderr = io.Discard
	if d1.Run() != nil || d2.Run() != nil {
		dirty = "*"
	}
	return fmt.Sprintf(" (%s%s)", branch, dirty)
}

func fmtReset(unix int64) string {
	if unix == 0 {
		return ""
	}
	remaining := time.Until(time.Unix(unix, 0))
	if remaining <= 0 {
		return ""
	}
	h := int(remaining.Hours())
	m := int(remaining.Minutes()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}

func fmtWinK(n int64) string {
	if n >= 1000 {
		return fmt.Sprintf("%dk", n/1000)
	}
	return fmt.Sprintf("%d", n)
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println(version)
		return
	}

	data, err := io.ReadAll(os.Stdin)
	if err != nil || len(strings.TrimSpace(string(data))) == 0 {
		fmt.Println()
		return
	}

	var in statusInput
	_ = json.Unmarshal(data, &in)

	cwd := in.Workspace.CurrentDir
	if cwd == "" {
		cwd = in.Cwd
	}
	if cwd == "" {
		cwd, _ = os.Getwd()
	}

	sep := dim + " | " + reset
	var out strings.Builder

	cwdShort := shortenPath(cwd)
	gitStr := gitInfo(cwd)
	model := in.Model.DisplayName
	if model == "" {
		model = in.Model.ID
	}

	out.WriteString(lightBlue + cwdShort + reset)
	out.WriteString(gitStr)
	if model != "" {
		out.WriteString(" " + dim + model + reset)
	}
	if in.ContextWindow.ContextWindowSize > 0 {
		pct := int(math.Round(in.ContextWindow.UsedPercentage))
		usedK := in.ContextWindow.TotalInputTokens / 1000
		winK := fmtWinK(in.ContextWindow.ContextWindowSize)
		color := ""
		if pct > 75 {
			color = yellow
		}
		out.WriteString(sep + fmt.Sprintf("%sctx: %d/%s (%d%%)%s", color, usedK, winK, pct, reset))
	}

	if in.RateLimits.FiveHour.UsedPercentage > 0 || in.RateLimits.SevenDay.UsedPercentage > 0 {
		out.WriteString(dim + " · " + reset)
		if in.RateLimits.FiveHour.UsedPercentage > 0 {
			s := int(math.Round(in.RateLimits.FiveHour.UsedPercentage))
			color := ""
			if s >= 80 {
				color = yellow
			}
			out.WriteString(fmt.Sprintf("%ssession: %d%%%s", color, s, reset))
		}
		if in.RateLimits.SevenDay.UsedPercentage > 0 {
			w := int(math.Round(in.RateLimits.SevenDay.UsedPercentage))
			color := ""
			if w >= 80 {
				color = yellow
			}
			out.WriteString(dim + " · " + reset + fmt.Sprintf("%sweek: %d%%%s", color, w, reset))
		}
	}

	if r := fmtReset(in.RateLimits.FiveHour.ResetsAt); r != "" {
		out.WriteString(dim + " · resets in " + r + reset)
	}

	out.WriteString("\n")
	fmt.Print(out.String())
}
