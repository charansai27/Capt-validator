package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// ValidationError represents a failed validation
type ValidationError struct {
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Details     interface{} `json:"details,omitempty"`
}

// Caption struct
type Caption struct {
	Start float64
	End   float64
	Text  string
}

func main() {
	tStart := flag.Float64("t_start", 0, "Start time in seconds")
	tEnd := flag.Float64("t_end", 0, "End time in seconds")
	coverage := flag.Float64("coverage", 0, "Coverage percentage (0-100)")
	endpoint := flag.String("endpoint", "", "Language detection endpoint URL (required)")
	flag.Parse()

	if *endpoint == "" {
		fmt.Fprintln(os.Stderr, "ERROR: -endpoint is required")
		os.Exit(1)
	}

	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "ERROR: captions file path is required")
		os.Exit(1)
	}

	filePath := args[0]
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != ".srt" && ext != ".vtt" {
		fmt.Fprintln(os.Stderr, "ERROR: unsupported file type")
		os.Exit(1)
	}

	captions, err := parseCaptions(filePath, ext)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR parsing captions:", err)
		os.Exit(1)
	}

	// Coverage check
	covered := calcCoverage(captions, *tStart, *tEnd)
	if covered < (*coverage / 100) {
		printJSON(ValidationError{
			Type:        "caption_coverage",
			Description: fmt.Sprintf("Coverage too low: %.2f%% < required %.2f%%", covered*100, *coverage),
		})
	}

	// Language check
	text := strings.Join(extractText(captions), "\n")
	lang, err := detectLanguage(*endpoint, text)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR detecting language:", err)
	} else if lang != "en-US" {
		printJSON(ValidationError{
			Type:        "incorrect_language",
			Description: fmt.Sprintf("Detected language %s is not acceptable", lang),
		})
	}
}

func parseCaptions(path, ext string) ([]Caption, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var captions []Caption
	var text []string
	var start, end float64

	srtTime := regexp.MustCompile(`(\d{2}):(\d{2}):(\d{2}),(\d+) --> (\d{2}):(\d{2}):(\d{2}),(\d+)`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			if text != nil {
				captions = append(captions, Caption{Start: start, End: end, Text: strings.Join(text, " ")})
				text = nil
			}
			continue
		}
		if ext == ".srt" && srtTime.MatchString(line) {
			m := srtTime.FindStringSubmatch(line)
			start = hmsToSeconds(m[1], m[2], m[3], m[4])
			end = hmsToSeconds(m[5], m[6], m[7], m[8])
			continue
		}
		if ext == ".vtt" && strings.Contains(line, "-->") {
			parts := strings.Split(line, " --> ")
			start, _ = parseVTTTime(parts[0])
			end, _ = parseVTTTime(parts[1])
			continue
		}
		text = append(text, line)
	}
	if text != nil {
		captions = append(captions, Caption{Start: start, End: end, Text: strings.Join(text, " ")})
	}
	return captions, scanner.Err()
}

func hmsToSeconds(h, m, s, ms string) float64 {
	hh, _ := strconv.Atoi(h)
	mm, _ := strconv.Atoi(m)
	ss, _ := strconv.Atoi(s)
	msec, _ := strconv.Atoi(ms)
	return float64(hh*3600+mm*60+ss) + float64(msec)/1000
}

func parseVTTTime(s string) (float64, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid VTT time: %s", s)
	}
	h, _ := strconv.Atoi(parts[0])
	m, _ := strconv.Atoi(parts[1])
	secParts := strings.Split(parts[2], ".")
	sec, _ := strconv.Atoi(secParts[0])
	ms := 0
	if len(secParts) > 1 {
		ms, _ = strconv.Atoi(secParts[1])
	}
	return float64(h*3600 + m*60 + sec + ms/1000), nil
}

func calcCoverage(captions []Caption, tStart, tEnd float64) float64 {
	total := tEnd - tStart
	if total <= 0 {
		return 0
	}
	covered := 0.0
	for _, c := range captions {
		s := max(c.Start, tStart)
		e := min(c.End, tEnd)
		if e > s {
			covered += e - s
		}
	}
	return covered / total
}

func extractText(captions []Caption) []string {
	var texts []string
	for _, c := range captions {
		texts = append(texts, c.Text)
	}
	return texts
}

func detectLanguage(endpoint, text string) (string, error) {
	resp, err := http.Post(endpoint, "text/plain", strings.NewReader(text))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var res struct {
		Lang string `json:"lang"`
	}
	b, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(b, &res); err != nil {
		return "", err
	}
	return res.Lang, nil
}

func printJSON(v ValidationError) {
	j, _ := json.Marshal(v)
	fmt.Println(string(j))
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
