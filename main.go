package main

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"
	"unicode"

	"github.com/espresso20/acronomicon/internal/data"
	"github.com/espresso20/acronomicon/internal/ui"
	"github.com/espresso20/acronomicon/internal/updater"
)

var version = "dev" // set via -ldflags "-X main.version=..."

func main() {
	initTerminal()

	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		ui.ColorDisabled = true
	}

	// Strip --no-color flag from args
	var args []string
	for _, a := range os.Args[1:] {
		if a == "--no-color" || a == "-no-color" {
			ui.ColorDisabled = true
			continue
		}
		args = append(args, a)
	}

	if len(args) == 0 {
		runREPL()
		return
	}

	cmd := strings.ToLower(args[0])
	switch cmd {
	case "help", "-h", "--help", "/?":
		ui.PrintBanner(len(data.All()))
		ui.PrintHelp()

	case "search", "-s", "--search", "find":
		if len(args) < 2 {
			fmt.Println(ui.C(ui.FgBrightRed, "Usage: acronomicon search <query>"))
			os.Exit(1)
		}
		query := strings.Join(args[1:], " ")
		handleSearch(query)

	case "list", "-l", "--list", "ls":
		if len(args) < 2 {
			ui.PrintAlphabetSummary()
		} else {
			handleListLetter(args[1])
		}

	case "category", "cat", "categories", "-c", "--category":
		if len(args) < 2 {
			ui.PrintCategories()
		} else {
			catQuery := strings.Join(args[1:], " ")
			handleCategory(catQuery)
		}

	case "random", "rand", "verse", "-r", "--random":
		fmt.Println()
		fmt.Println(ui.C(ui.FgBrightMagenta+ui.Bold, "✨ ACRONOMICON VERSE OF THE DAY ✨"))
		ui.PrintCard(data.Random())

	case "quiz", "-q", "--quiz":
		reader := bufio.NewReader(os.Stdin)
		ui.RunQuiz(reader, 5)

	case "stats", "stat", "--stats":
		ui.PrintStats()

	case "update", "-u", "--update", "upgrade":
		os.Exit(runUpdate())

	case "version", "-v", "--version":
		fmt.Println("acronomicon " + version)

	default:
		// If single letter A-Z, show letter list OR exact acronym if desired;
		// for direct CLI 'acronomicon k8s' or 'acronomicon vpc', look up the acronym(s).
		query := strings.Join(args, " ")
		handleLookup(query)
	}
}

func handleLookup(query string) {
	exact := data.Lookup(query)
	if len(exact) > 0 {
		for _, e := range exact {
			ui.PrintCard(e)
		}
		return
	}

	// If single letter A-Z and no exact match, show letter list
	trimmed := strings.TrimSpace(query)
	if len(trimmed) == 1 && unicode.IsLetter(rune(trimmed[0])) {
		handleListLetter(trimmed)
		return
	}

	fmt.Println()
	fmt.Printf("  %s No exact acronym match for %s.\n",
		ui.C(ui.FgBrightYellow+ui.Bold, "⚠"),
		ui.C(ui.FgBrightWhite+ui.Bold, fmt.Sprintf("%q", query)),
	)

	suggestions := data.Suggest(query, 8)
	if len(suggestions) > 0 {
		ui.PrintTable(fmt.Sprintf("Did you mean one of these matches for %q?", query), suggestions)
		return
	}

	// Fall back to full-text search
	searchRes := data.Search(query)
	if len(searchRes) > 0 {
		if len(searchRes) > 12 {
			searchRes = searchRes[:12]
		}
		ui.PrintTable(fmt.Sprintf("Search results mentioning %q", query), searchRes)
		return
	}

	fmt.Printf("  %s Try %s or %s to browse all %d acronyms.\n",
		ui.C(ui.Dim, "💡"),
		ui.C(ui.FgBrightCyan, "acronomicon list"),
		ui.C(ui.FgBrightCyan, "acronomicon search <keyword>"),
		len(data.All()),
	)
}

func handleSearch(query string) {
	results := data.Search(query)
	if len(results) == 0 {
		fmt.Printf("\n  %s No results found matching %q.\n", ui.C(ui.FgBrightRed, "✗"), query)
		return
	}
	ui.PrintTable(fmt.Sprintf("Search Results for %q", query), results)
}

func handleListLetter(letter string) {
	letter = strings.TrimSpace(letter)
	if len(letter) != 1 || !unicode.IsLetter(rune(letter[0])) {
		ui.PrintAlphabetSummary()
		return
	}
	ch := unicode.ToUpper(rune(letter[0]))
	entries := data.ByLetter(string(ch))
	ui.PrintTable(fmt.Sprintf("Acronyms Starting with '%c'", ch), entries)
}

func handleCategory(catQuery string) {
	matchedName, entries := data.ByCategory(catQuery)
	if len(entries) == 0 {
		fmt.Printf("\n  %s No category matching %q.\n", ui.C(ui.FgBrightRed, "✗"), catQuery)
		ui.PrintCategories()
		return
	}
	ui.PrintTable("Category: "+matchedName, entries)
}

func runUpdate() int {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	u := updater.New(version)
	fmt.Println(ui.C(ui.FgBrightCyan+ui.Bold, "⟳ Checking for updates..."))

	err := u.Update(ctx)
	if err != nil {
		fmt.Println(ui.C(ui.FgBrightRed, "✗ Update failed: "+err.Error()))
		return 1
	}
	return 0
}

func runREPL() {
	total := len(data.All())
	ui.PrintBanner(total)
	fmt.Println()
	fmt.Printf("  Welcome to %s! Type any acronym (e.g. %s, %s, %s, %s),\n",
		ui.C(ui.FgBrightYellow+ui.Bold, "The Acronomicon"),
		ui.C(ui.FgBrightCyan+ui.Bold, "k8s"),
		ui.C(ui.FgBrightCyan+ui.Bold, "vpc"),
		ui.C(ui.FgBrightCyan+ui.Bold, "ebpf"),
		ui.C(ui.FgBrightCyan+ui.Bold, "mtls"),
	)
	fmt.Printf("  a single letter (%s–%s), %s, %s, %s, %s, %s, or %s to exit.\n",
		ui.C(ui.FgBrightGreen, "a"),
		ui.C(ui.FgBrightGreen, "z"),
		ui.C(ui.FgBrightGreen, "?<search>"),
		ui.C(ui.FgBrightGreen, "/cat"),
		ui.C(ui.FgBrightGreen, "/random"),
		ui.C(ui.FgBrightGreen, "/quiz"),
		ui.C(ui.FgBrightGreen, "/help"),
		ui.C(ui.FgBrightRed, "q"),
	)

	var updateReady atomic.Bool
	if os.Getenv("ACRONOMICON_NO_UPDATE_CHECK") == "" {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			u := updater.New(version)
			u.Client = &http.Client{Timeout: 5 * time.Second}
			if st, err := u.CheckCached(ctx, 24*time.Hour); err == nil && st.Available {
				updateReady.Store(true)
			}
		}()
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		if updateReady.CompareAndSwap(true, false) {
			fmt.Printf("\n  %s\n", ui.C(ui.Dim, "↑ A new version of The Acronomicon is available — type /update to install."))
		}

		fmt.Printf("\n%s ", ui.C(ui.FgBrightCyan+ui.Bold, "acronomicon>"))
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println()
			break
		}
		input := strings.TrimSpace(line)
		if input == "" {
			continue
		}

		lower := strings.ToLower(input)
		switch {
		case lower == "q" || lower == "quit" || lower == "exit" || lower == "/q" || lower == "/quit":
			fmt.Println(ui.C(ui.FgBrightYellow, "May your latencies be low and your SLAs five nines. Farewell! 📖✨"))
			return

		case lower == "?" || lower == "help" || lower == "/help" || lower == "/h":
			ui.PrintHelp()

		case lower == "/r" || lower == "/random" || lower == "random":
			fmt.Println()
			fmt.Println(ui.C(ui.FgBrightMagenta+ui.Bold, "✨ ACRONOMICON VERSE OF THE DAY ✨"))
			ui.PrintCard(data.Random())

		case lower == "/list" || lower == "/ls" || lower == "list":
			ui.PrintAlphabetSummary()

		case lower == "/stats" || lower == "stats":
			ui.PrintStats()

		case lower == "/quiz" || lower == "quiz":
			ui.RunQuiz(reader, 5)

		case lower == "/cat" || lower == "/categories" || lower == "categories":
			ui.PrintCategories()

		case lower == "/update" || lower == "update":
			runUpdate()

		case strings.HasPrefix(lower, "/cat ") || strings.HasPrefix(lower, "/category "):
			parts := strings.SplitN(input, " ", 2)
			if len(parts) == 2 {
				handleCategory(parts[1])
			} else {
				ui.PrintCategories()
			}

		case strings.HasPrefix(lower, "/search ") || strings.HasPrefix(lower, "/s "):
			parts := strings.SplitN(input, " ", 2)
			if len(parts) == 2 {
				handleSearch(parts[1])
			}

		case strings.HasPrefix(input, "?") && len(input) > 1:
			handleSearch(strings.TrimSpace(input[1:]))

		case len(input) == 1 && unicode.IsLetter(rune(input[0])):
			handleListLetter(input)

		default:
			handleLookup(input)
		}
	}
}
