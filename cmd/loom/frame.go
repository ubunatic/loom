package main

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"ubunatic.com/loom/measure"
)

type frameStyle struct{ topLeft, topRight, bottomLeft, bottomRight, horizontal, vertical rune }

var frameStyles = map[string]frameStyle{
	"single":  {'┌', '┐', '└', '┘', '─', '│'},
	"double":  {'╔', '╗', '╚', '╝', '═', '║'},
	"rounded": {'╭', '╮', '╰', '╯', '─', '│'},
	"heavy":   {'┏', '┓', '┗', '┛', '━', '┃'},
}

func frameCommand() *cobra.Command {
	var style, title, color string
	var padding int
	var write bool
	cmd := &cobra.Command{Use: "frame <file>", Short: "Wrap text in a Unicode box frame", Args: cobra.ExactArgs(1), SilenceUsage: true}
	cmd.Flags().StringVar(&style, "style", "single", "border style: single, double, rounded, heavy")
	cmd.Flags().StringVar(&title, "title", "", "centered title in the top border")
	cmd.Flags().IntVar(&padding, "padding", 0, "horizontal and vertical content padding")
	cmd.Flags().StringVar(&color, "color", "", "border foreground color name or palette index")
	cmd.Flags().BoolVarP(&write, "write", "w", false, "replace the file atomically")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if padding < 0 {
			return fmt.Errorf("padding must be non-negative")
		}
		data, err := readAsset(args[0])
		if err != nil {
			return err
		}
		framed, err := frameText(string(data), style, title, padding, color)
		if err != nil {
			return err
		}
		if write {
			return atomicWrite(args[0], []byte(framed))
		}
		_, err = io.WriteString(cmd.OutOrStdout(), framed)
		return err
	}
	return cmd
}

func frameText(text, styleName, title string, padding int, color string) (string, error) {
	style, ok := frameStyles[styleName]
	if !ok {
		return "", fmt.Errorf("unknown frame style %q", styleName)
	}
	if padding < 0 {
		return "", fmt.Errorf("padding must be non-negative")
	}
	colorCode, err := parseBorderColor(color)
	if err != nil {
		return "", err
	}
	text = strings.ReplaceAll(text, "\r\n", "\n")
	finalNL := strings.HasSuffix(text, "\n")
	lines := strings.Split(text, "\n")
	if finalNL {
		lines = lines[:len(lines)-1]
	}
	maxWidth := 0
	for _, line := range lines {
		if w := measure.StringWidth(line); w > maxWidth {
			maxWidth = w
		}
	}
	inner := maxWidth + 2*padding
	if inner < 1 {
		inner = 1
	}
	if title != "" && measure.StringWidth(title)+4 > inner {
		inner = measure.StringWidth(title) + 4
	}
	var out []string
	top := ""
	if title == "" {
		top = colorizeBorder(string(style.topLeft)+strings.Repeat(string(style.horizontal), inner)+string(style.topRight), colorCode)
	} else {
		titleW := measure.StringWidth(title)
		remaining := inner - titleW - 2
		left := remaining / 2
		right := remaining - left
		leftRun := string(style.topLeft) + strings.Repeat(string(style.horizontal), left)
		rightRun := strings.Repeat(string(style.horizontal), right) + string(style.topRight)
		top = colorizeBorder(leftRun, colorCode) + " " + title + " " + colorizeBorder(rightRun, colorCode)
	}
	out = append(out, top)
	blank := strings.Repeat(" ", inner)
	for i := 0; i < padding; i++ {
		out = append(out, colorizeBorder(string(style.vertical), colorCode)+blank+colorizeBorder(string(style.vertical), colorCode))
	}
	for _, line := range lines {
		w := measure.StringWidth(line)
		left := padding
		right := inner - w - left
		if right < 0 {
			right = 0
		}
		out = append(out, colorizeBorder(string(style.vertical), colorCode)+strings.Repeat(" ", left)+line+strings.Repeat(" ", right)+colorizeBorder(string(style.vertical), colorCode))
	}
	for i := 0; i < padding; i++ {
		out = append(out, colorizeBorder(string(style.vertical), colorCode)+blank+colorizeBorder(string(style.vertical), colorCode))
	}
	bottom := colorizeBorder(string(style.bottomLeft)+strings.Repeat(string(style.horizontal), inner)+string(style.bottomRight), colorCode)
	out = append(out, bottom)
	result := strings.Join(out, "\n")
	if finalNL {
		result += "\n"
	}
	return result, nil
}

func parseBorderColor(name string) (string, error) {
	if name == "" {
		return "", nil
	}
	lower := strings.ToLower(strings.TrimSpace(name))
	basic := map[string]int{"black": 30, "red": 31, "green": 32, "yellow": 33, "blue": 34, "magenta": 35, "cyan": 36, "white": 37, "gray": 90, "grey": 90}
	if code, ok := basic[lower]; ok {
		return fmt.Sprintf("%d", code), nil
	}
	if lower == "orange" {
		return "38;5;208", nil
	}
	index, err := strconv.Atoi(lower)
	if err != nil || index < 0 || index > 255 {
		return "", fmt.Errorf("invalid border color %q", name)
	}
	return fmt.Sprintf("38;5;%d", index), nil
}

func colorizeBorder(s, code string) string {
	if code == "" {
		return s
	}
	return "\x1b[" + code + "m" + s + "\x1b[0m"
}
