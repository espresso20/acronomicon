package ui

import (
	"bufio"
	"fmt"
	"math/rand"
	"strings"
	"unicode/utf8"

	"github.com/espresso20/acronomicon/internal/data"
)

// ColorDisabled controls whether ANSI escape codes are stripped.
var ColorDisabled = false

const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Italic    = "\033[3m"
	Underline = "\033[4m"

	FgBlack   = "\033[30m"
	FgRed     = "\033[31m"
	FgGreen   = "\033[32m"
	FgYellow  = "\033[33m"
	FgBlue    = "\033[34m"
	FgMagenta = "\033[35m"
	FgCyan    = "\033[36m"
	FgWhite   = "\033[37m"

	FgBrightRed     = "\033[91m"
	FgBrightGreen   = "\033[92m"
	FgBrightYellow  = "\033[93m"
	FgBrightBlue    = "\033[94m"
	FgBrightMagenta = "\033[95m"
	FgBrightCyan    = "\033[96m"
	FgBrightWhite   = "\033[97m"

	BgBlue    = "\033[44m"
	BgMagenta = "\033[45m"
	BgCyan    = "\033[46m"
	BgYellow  = "\033[43m"
	BgGreen   = "\033[42m"
	BgRed     = "\033[41m"
)

// C wraps text in ANSI codes unless ColorDisabled is true.
func C(code, text string) string {
	if ColorDisabled {
		return text
	}
	return code + text + Reset
}

// Badge renders a pill-like colored tag.
func Badge(bg, fg, text string) string {
	if ColorDisabled {
		return "[" + text + "]"
	}
	return bg + fg + Bold + " " + text + " " + Reset
}

func providerBadge(provider string) string {
	switch strings.ToUpper(provider) {
	case "AWS":
		return Badge(BgYellow, FgBlack, "AWS")
	case "GCP":
		return Badge(BgBlue, FgBrightWhite, "GCP")
	case "AZURE":
		return Badge(BgCyan, FgBlack, "Azure")
	case "CNCF":
		return Badge(BgMagenta, FgBrightWhite, "CNCF")
	default:
		return Badge(BgGreen, FgBlack, provider)
	}
}

func categoryColor(cat string) string {
	switch {
	case strings.Contains(cat, "Network"):
		return FgBrightCyan
	case strings.Contains(cat, "Security"):
		return FgBrightRed
	case strings.Contains(cat, "Container"):
		return FgBrightMagenta
	case strings.Contains(cat, "Storage"):
		return FgBrightYellow
	case strings.Contains(cat, "Compute"):
		return FgBrightGreen
	case strings.Contains(cat, "DevOps"):
		return FgBrightBlue
	case strings.Contains(cat, "Observability"):
		return FgCyan
	case strings.Contains(cat, "Serverless"):
		return FgYellow
	default:
		return FgBrightWhite
	}
}

// visibleLen calculates the printable rune width of a string, ignoring ANSI escape sequences.
func visibleLen(s string) int {
	inEsc := false
	count := 0
	for _, r := range s {
		if r == '\033' {
			inEsc = true
			continue
		}
		if inEsc {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEsc = false
			}
			continue
		}
		count++
	}
	return count
}

// padRight pads a string (which may contain ANSI codes) with spaces to reach width visible runes.
func padRight(s string, width int) string {
	vLen := visibleLen(s)
	if vLen >= width {
		return s
	}
	return s + strings.Repeat(" ", width-vLen)
}

// wrapText word-wraps plain text to maxWidth runes per line.
func wrapText(text string, maxWidth int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}
	var lines []string
	current := words[0]
	for _, w := range words[1:] {
		if utf8.RuneCountInString(current)+1+utf8.RuneCountInString(w) <= maxWidth {
			current += " " + w
		} else {
			lines = append(lines, current)
			current = w
		}
	}
	lines = append(lines, current)
	return lines
}

// PrintBanner renders the startup banner for THE ACRONOMICON.
func PrintBanner(totalCount int) {
	border := C(FgBrightCyan+Bold, "╔═══════════════════════════════════════════════════════════════════════════════╗")
	bottom := C(FgBrightCyan+Bold, "╚═══════════════════════════════════════════════════════════════════════════════╝")
	side := C(FgBrightCyan+Bold, "║")

	art := []string{
		"   _____                                             .__                     ",
		"  /  _  \\   ____ _______  ____   ____   ____   _____ |__| ____  ____   ____  ",
		" /  /_\\  \\_/ ___\\\\_  __ \\/  _ \\ /    \\ /  _ \\ /     \\|  |/ ___\\/  _ \\ /    \\ ",
		"/    |    \\  \\___ |  | \\(  <_> )   |  (  <_> )  Y Y  \\  \\  \\__(  <_> )   |  \\",
		"\\____|__  /\\___  >|__|   \\____/|___|  /\\____/|__|_|  /__|\\___  >____/|___|  /",
		"        \\/     \\/                   \\/             \\/        \\/           \\/ ",
	}

	fmt.Println()
	fmt.Println(border)
	for _, line := range art {
		padded := padRight(C(FgBrightYellow+Bold, line), 77)
		fmt.Printf("%s %s %s\n", side, padded, side)
	}
	sub := fmt.Sprintf("        ✦  THE BIBLE OF CLOUD TECH ACRONYMS • %d+ Entries (A–Z)  ✦", totalCount)
	fmt.Printf("%s %s %s\n", side, padRight(C(FgBrightWhite+Bold, sub), 77), side)
	fmt.Println(bottom)
}

// PrintCard renders a full-color detailed card for an acronym Entry.
func PrintCard(e data.Entry) {
	const boxWidth = 78 // inner width between │ and │
	const textWidth = 74

	topBorder := C(FgBrightCyan, "╭"+strings.Repeat("─", boxWidth)+"╮")
	midBorder := C(FgBrightCyan, "├"+strings.Repeat("─", boxWidth)+"┤")
	botBorder := C(FgBrightCyan, "╰"+strings.Repeat("─", boxWidth)+"╯")
	vBar := C(FgBrightCyan, "│")

	printRow := func(content string) {
		fmt.Printf("%s  %s  %s\n", vBar, padRight(content, textWidth), vBar)
	}

	fmt.Println()
	fmt.Println(topBorder)

	// Header row: Acronym badge + Provider badge + Category tag
	acrBadge := Badge(BgCyan, FgBlack, " "+e.Acronym+" ")
	prov := providerBadge(e.Provider)
	catTag := C(categoryColor(e.Category)+Bold, "["+e.Category+"]")

	aliasStr := ""
	if len(e.Aliases) > 0 {
		aliasStr = C(Dim+FgWhite, " (aka "+strings.Join(e.Aliases, ", ")+")")
	}

	headerLine := fmt.Sprintf("%s%s   %s  %s", acrBadge, aliasStr, prov, catTag)
	printRow(headerLine)

	// Full Name row
	fullLines := wrapText(e.FullName, textWidth-2)
	for i, l := range fullLines {
		prefix := "✦ "
		if i > 0 {
			prefix = "  "
		}
		printRow(C(FgBrightYellow+Bold, prefix+l))
	}

	fmt.Println(midBorder)

	// Meaning section
	printRow(C(FgBrightGreen+Bold+Underline, "MEANING (TL;DR)"))
	for _, l := range wrapText(e.Meaning, textWidth) {
		printRow(C(FgBrightWhite, l))
	}
	printRow("")

	// What It Does section
	printRow(C(FgBrightCyan+Bold+Underline, "WHAT IT DOES & WHY IT MATTERS"))
	for _, l := range wrapText(e.WhatItDoes, textWidth) {
		printRow(C(FgWhite, l))
	}
	printRow("")

	// Practical Example section
	printRow(C(FgBrightMagenta+Bold+Underline, "IN PRACTICE / REAL-WORLD EXAMPLE"))
	for _, l := range wrapText(e.Example, textWidth-2) {
		printRow(C(FgBrightYellow, "▸ ") + C(Italic+FgWhite, l))
	}

	// Related section
	if len(e.Related) > 0 {
		fmt.Println(midBorder)
		var relPills []string
		for _, r := range e.Related {
			relPills = append(relPills, C(FgBrightCyan+Bold, r))
		}
		relLine := C(Dim+FgWhite, "See Also: ") + strings.Join(relPills, C(Dim, " • "))
		printRow(relLine)
	}

	fmt.Println(botBorder)
}

// PrintTable renders a compact colored list of entries (for search, letter list, or category list).
func PrintTable(title string, entries []data.Entry) {
	fmt.Println()
	fmt.Printf("%s %s\n", C(FgBrightCyan+Bold, "━━━"), C(FgBrightYellow+Bold, title)+C(Dim, fmt.Sprintf(" (%d entries)", len(entries))))
	fmt.Println(C(FgCyan, strings.Repeat("─", 80)))
	fmt.Printf("  %-14s %-15s %-25s %s\n",
		C(Bold+FgBrightWhite, "ACRONYM"),
		C(Bold+FgBrightWhite, "PROVIDER"),
		C(Bold+FgBrightWhite, "CATEGORY"),
		C(Bold+FgBrightWhite, "FULL NAME"),
	)
	fmt.Println(C(FgCyan, strings.Repeat("─", 80)))

	for _, e := range entries {
		acr := padRight(C(FgBrightCyan+Bold, e.Acronym), 14)
		prov := padRight(C(FgYellow, e.Provider), 15)
		cat := padRight(C(categoryColor(e.Category), truncate(e.Category, 24)), 25)
		full := truncate(e.FullName, 44)
		fmt.Printf("  %s %s %s %s\n", acr, prov, cat, C(FgBrightWhite, full))
	}
	fmt.Println(C(FgCyan, strings.Repeat("─", 80)))
	fmt.Printf("  %s\n", C(Dim, "💡 Tip: Type 'acronomicon <acronym>' (or just '<acronym>' in interactive mode) for full details."))
}

// PrintAlphabetSummary prints all entries grouped by A-Z letter.
func PrintAlphabetSummary() {
	all := data.All()
	grouped := make(map[rune][]string)
	for _, e := range all {
		grouped[e.FirstLetter()] = append(grouped[e.FirstLetter()], e.Acronym)
	}

	fmt.Println()
	fmt.Println(C(FgBrightYellow+Bold, "📖 THE ACRONOMICON INDEX (A–Z)") + C(Dim, fmt.Sprintf(" — %d Total Cloud Acronyms", len(all))))
	fmt.Println(C(FgBrightCyan, strings.Repeat("═", 80)))

	for ch := 'A'; ch <= 'Z'; ch++ {
		list := grouped[ch]
		letterBadge := Badge(BgCyan, FgBlack, fmt.Sprintf("%c", ch))
		countStr := C(Dim, fmt.Sprintf("(%2d)", len(list)))
		joined := strings.Join(list, ", ")
		lines := wrapText(joined, 66)
		for i, line := range lines {
			if i == 0 {
				fmt.Printf(" %s %s  %s\n", letterBadge, countStr, C(FgBrightWhite, line))
			} else {
				fmt.Printf("          %s\n", C(FgBrightWhite, line))
			}
		}
	}
	fmt.Println(C(FgBrightCyan, strings.Repeat("═", 80)))
	fmt.Println(C(Dim, " 💡 Run 'acronomicon list <letter>' (e.g. 'acronomicon list k') to inspect a specific letter."))
}

// PrintCategories prints all available categories and their entry counts.
func PrintCategories() {
	cats := data.Categories()
	fmt.Println()
	fmt.Println(C(FgBrightYellow+Bold, "🗂️  ACRONOMICON CATEGORIES"))
	fmt.Println(C(FgBrightCyan, strings.Repeat("─", 66)))
	for i, c := range cats {
		barLen := c.Count / 2
		if barLen < 1 {
			barLen = 1
		}
		bar := C(categoryColor(c.Name), strings.Repeat("█", barLen))
		num := C(Dim, fmt.Sprintf("%2d.", i+1))
		name := padRight(C(categoryColor(c.Name)+Bold, c.Name), 30)
		cnt := padRight(C(FgBrightWhite, fmt.Sprintf("%3d acronyms", c.Count)), 14)
		fmt.Printf("  %s %s %s %s\n", num, name, cnt, bar)
	}
	fmt.Println(C(FgBrightCyan, strings.Repeat("─", 66)))
	fmt.Println(C(Dim, " 💡 Run 'acronomicon category <name>' (e.g. 'acronomicon category security') to view entries."))
}

// PrintStats displays database breakdown across Categories, Providers, and A-Z distribution.
func PrintStats() {
	all := data.All()
	fmt.Println()
	fmt.Println(C(FgBrightYellow+Bold, "📊 THE ACRONOMICON DATABASE STATISTICS"))
	fmt.Println(C(FgBrightCyan, strings.Repeat("═", 72)))
	fmt.Printf("  %s %s\n", C(Bold+FgBrightWhite, "Total Cloud Acronyms:"), C(FgBrightGreen+Bold, fmt.Sprintf("%d", len(all))))

	PrintCategories()

	fmt.Println()
	fmt.Println(C(FgBrightYellow+Bold, "☁️  BY CLOUD PROVIDER / ECOSYSTEM"))
	fmt.Println(C(FgBrightCyan, strings.Repeat("─", 66)))
	for _, p := range data.Providers() {
		pct := float64(p.Count) * 100.0 / float64(len(all))
		barLen := int(pct / 2.5)
		if barLen < 1 {
			barLen = 1
		}
		fmt.Printf("  %-20s %4d (%5.1f%%)  %s\n",
			padRight(providerBadge(p.Name), 20),
			p.Count,
			pct,
			C(FgBrightCyan, strings.Repeat("▓", barLen)),
		)
	}

	fmt.Println()
	fmt.Println(C(FgBrightYellow+Bold, "🔤 ALPHABET DISTRIBUTION (A–Z)"))
	fmt.Println(C(FgBrightCyan, strings.Repeat("─", 66)))
	letters := make(map[rune]int)
	for _, e := range all {
		letters[e.FirstLetter()]++
	}
	for row := 0; row < 13; row++ {
		ch1 := rune('A' + row)
		ch2 := rune('A' + row + 13)
		c1 := letters[ch1]
		c2 := letters[ch2]
		b1 := C(FgBrightGreen, strings.Repeat("▪", minInt(c1, 20)))
		b2 := C(FgBrightGreen, strings.Repeat("▪", minInt(c2, 20)))
		left := fmt.Sprintf("  %s (%2d) %-20s", C(FgBrightCyan+Bold, string(ch1)), c1, b1)
		right := fmt.Sprintf("  %s (%2d) %-20s", C(FgBrightCyan+Bold, string(ch2)), c2, b2)
		fmt.Printf("%-38s %s\n", padRight(left, 34), right)
	}
	fmt.Println(C(FgBrightCyan, strings.Repeat("═", 72)))
}

// RunQuiz runs an interactive multiple-choice quiz in the terminal.
func RunQuiz(reader *bufio.Reader, rounds int) {
	all := data.All()
	if len(all) < 4 {
		return
	}
	fmt.Println()
	fmt.Println(C(FgBrightYellow+Bold, "⚡ THE ACRONOMICON CLOUD MASTER QUIZ ⚡"))
	fmt.Println(C(Dim, fmt.Sprintf("Answer %d multiple-choice questions (1–4, or 'q' to stop early).", rounds)))

	score := 0
	asked := 0

	perm := rand.Perm(len(all))
	for r := 0; r < rounds && r < len(all); r++ {
		target := all[perm[r]]

		// Pick 3 distractors
		options := []data.Entry{target}
		for len(options) < 4 {
			cand := all[rand.Intn(len(all))]
			dup := false
			for _, o := range options {
				if o.Acronym == cand.Acronym {
					dup = true
					break
				}
			}
			if !dup {
				options = append(options, cand)
			}
		}
		rand.Shuffle(len(options), func(i, j int) {
			options[i], options[j] = options[j], options[i]
		})

		correctIdx := 0
		for i, o := range options {
			if o.Acronym == target.Acronym {
				correctIdx = i + 1
				break
			}
		}

		fmt.Println()
		fmt.Printf("%s What does %s (%s) stand for and do?\n",
			C(FgBrightCyan+Bold, fmt.Sprintf("[Q%d/%d]", r+1, rounds)),
			Badge(BgCyan, FgBlack, " "+target.Acronym+" "),
			C(FgYellow, target.Category),
		)
		for i, o := range options {
			fmt.Printf("  %s %s — %s\n",
				C(FgBrightYellow+Bold, fmt.Sprintf("[%d]", i+1)),
				C(Bold+FgBrightWhite, o.FullName),
				C(Dim+FgWhite, truncate(o.Meaning, 68)),
			)
		}

		fmt.Printf("\n%s ", C(FgBrightCyan+Bold, "Your Answer (1-4, or q):"))
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		ans := strings.TrimSpace(strings.ToLower(line))
		if ans == "q" || ans == "quit" || ans == "exit" {
			break
		}

		asked++
		if ans == fmt.Sprintf("%d", correctIdx) {
			score++
			fmt.Printf("  %s %s stands for %s!\n",
				C(FgBrightGreen+Bold, "✅ CORRECT!"),
				C(FgBrightCyan+Bold, target.Acronym),
				C(Bold, target.FullName),
			)
		} else {
			fmt.Printf("  %s The correct answer was [%d]: %s (%s)\n",
				C(FgBrightRed+Bold, "❌ NOT QUITE!"),
				correctIdx,
				C(FgBrightYellow+Bold, target.FullName),
				target.Meaning,
			)
		}
	}

	if asked > 0 {
		fmt.Println()
		fmt.Println(C(FgBrightCyan, strings.Repeat("─", 60)))
		fmt.Printf("🏆 %s %s (%d%%)\n",
			C(FgBrightYellow+Bold, "Final Quiz Score:"),
			C(FgBrightGreen+Bold, fmt.Sprintf("%d / %d", score, asked)),
			(score*100)/asked,
		)
		fmt.Println(C(FgBrightCyan, strings.Repeat("─", 60)))
	}
}

// PrintHelp displays CLI and Interactive REPL usage instructions.
func PrintHelp() {
	fmt.Println()
	fmt.Println(C(FgBrightYellow+Bold, "📖 ACRONOMICON — USAGE & COMMANDS"))
	fmt.Println(C(FgBrightCyan, strings.Repeat("─", 74)))
	fmt.Println(C(Bold+FgBrightWhite, "Direct CLI Usage:"))
	fmt.Printf("  %-34s %s\n", C(FgBrightCyan, "acronomicon"), "Launch interactive REPL mode")
	fmt.Printf("  %-34s %s\n", C(FgBrightCyan, "acronomicon <acronym>"), "Look up an acronym card (e.g., vpc, k8s, ebpf)")
	fmt.Printf("  %-34s %s\n", C(FgBrightCyan, "acronomicon search <query>"), "Full-text search across names & descriptions (-s)")
	fmt.Printf("  %-34s %s\n", C(FgBrightCyan, "acronomicon list [A-Z]"), "List all acronyms or filter by first letter (-l)")
	fmt.Printf("  %-34s %s\n", C(FgBrightCyan, "acronomicon category [name]"), "List categories or acronyms in a category (-c)")
	fmt.Printf("  %-34s %s\n", C(FgBrightCyan, "acronomicon random"), "Display a random cloud acronym card (-r)")
	fmt.Printf("  %-34s %s\n", C(FgBrightCyan, "acronomicon quiz"), "Take a 5-question cloud acronym quiz")
	fmt.Printf("  %-34s %s\n", C(FgBrightCyan, "acronomicon stats"), "Show database statistics & A–Z breakdown")
	fmt.Printf("  %-34s %s\n", C(FgBrightCyan, "acronomicon --no-color ..."), "Disable ANSI colors (also respects NO_COLOR=1)")
	fmt.Println()
	fmt.Println(C(Bold+FgBrightWhite, "Interactive Prompt Commands (acronomicon>):"))
	fmt.Printf("  %-34s %s\n", C(FgBrightGreen, "<acronym>"), "Type any acronym (e.g. 'iam', 'alb', 'mtls')")
	fmt.Printf("  %-34s %s\n", C(FgBrightGreen, "a .. z"), "Type a single letter (a-z) to list that letter")
	fmt.Printf("  %-34s %s\n", C(FgBrightGreen, "/search <q>  or  ?<q>"), "Search all acronyms and descriptions")
	fmt.Printf("  %-34s %s\n", C(FgBrightGreen, "/cat [name]"), "Browse categories or filter by category name")
	fmt.Printf("  %-34s %s\n", C(FgBrightGreen, "/list"), "Print the complete A–Z index")
	fmt.Printf("  %-34s %s\n", C(FgBrightGreen, "/random  or  /r"), "Show a random 'Acronomicon Verse of the Day'")
	fmt.Printf("  %-34s %s\n", C(FgBrightGreen, "/quiz"), "Launch a quick 5-question multiple-choice quiz")
	fmt.Printf("  %-34s %s\n", C(FgBrightGreen, "/stats"), "View database stats and charts")
	fmt.Printf("  %-34s %s\n", C(FgBrightGreen, "/help  or  quit / q"), "Show help or exit The Acronomicon")
	fmt.Println(C(FgBrightCyan, strings.Repeat("─", 74)))
}

func truncate(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return string(runes[:maxRunes-1]) + "…"
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
