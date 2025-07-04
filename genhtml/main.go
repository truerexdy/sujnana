package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"encoding/json"
	"time"
	"strconv"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/joho/godotenv"
)

type Item struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	UpdatedTime string `json:"updated_time"`
}

func getNextID() (int, error) {
	lastIDStr := os.Getenv("LAST_ID")
	if lastIDStr == "" {
		return 0, nil
	}
	lastID, err := strconv.Atoi(lastIDStr)
	if err != nil {
		return 0, err
	}
	return lastID + 1, nil
}

func updateEnvFile(nextID int) error {
	envPath := "config.env"
	content, err := os.ReadFile(envPath)
	if err != nil {
		return err
	}

	lines := bytes.Split(content, []byte("\n"))
	var newLines [][]byte
	found := false

	for _, line := range lines {
		if bytes.HasPrefix(line, []byte("LAST_ID=")) {
			newLines = append(newLines, []byte(fmt.Sprintf("LAST_ID=%d", nextID)))
			found = true
		} else {
			newLines = append(newLines, line)
		}
	}

	if !found {
		newLines = append(newLines, []byte(fmt.Sprintf("LAST_ID=%d", nextID)))
	}

	newContent := bytes.Join(newLines, []byte("\n"))
	return os.WriteFile(envPath, newContent, 0644)
}

func updateJSONFile(filePath string, newItem Item) error {
	var items []Item

	data, err := os.ReadFile(filePath)
	if err == nil && len(data) > 0 {
		if err := json.Unmarshal(data, &items); err != nil {
			items = []Item{}
		}
	}

	items = append(items, newItem)

	updatedData, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %v", err)
	}

	err = os.WriteFile(filePath, updatedData, 0644)
	if err != nil {
		return fmt.Errorf("error writing file: %v", err)
	}
	return nil
}

func mdToHTML(md []byte) []byte {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)
	return markdown.Render(doc, renderer)
}

func injectHTML(b []byte, insert []byte, title []byte) []byte {
	idx := bytes.Index(b, []byte("<!--title-->"))
	if idx == -1 {
		return b
	}
	idx1 := bytes.Index(b, []byte("<!--body-->"))
	if idx1 == -1 {
		return b
	}
	out := make([]byte, 0, len(b)-len([]byte("0_title_0"))+len(title)+len(insert))
	out = append(out, b[:idx]...)
	out = append(out, title...)
	out = append(out, b[idx+len([]byte("<!--title-->")):idx1]...)
	out = append(out, insert...)
	out = append(out, b[idx1+len([]byte("<!--body-->")):]...)
	return out
}

func main() {
	err := godotenv.Load("config.env")
	if err != nil {
		fmt.Println("Couldn't load .env")
		return
	}

	htmlDir := os.Getenv("HTML_DIR")
	templatePath := os.Getenv("TEMPLATE_PATH")
	sourceDir := os.Getenv("SOURCE_DIR")
	if htmlDir == "" || templatePath == "" || sourceDir == "" {
		fmt.Println("Missing HTML_DIR, TEMPLATE_PATH, or SOURCE_DIR in env")
		return
	}

	fmt.Println("Enter markdown filename (without .md extension):")
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		fmt.Println("Couldn't read from stdin")
		return
	}
	filename := scanner.Text()
	
	fmt.Println("Enter title for the blog:")
	if !scanner.Scan() {
		fmt.Println("Couldn't read from stdin")
		return
	}
	title := scanner.Text()

	mdPath := filepath.Join(sourceDir, filename+".md")
	md, err := os.ReadFile(mdPath)
	if err != nil {
		fmt.Println("Could not read md file:", err)
		return
	}

	htmlTemplate, err := os.ReadFile(templatePath)
	if err != nil {
		fmt.Println("Could not read HTML template:", err)
		return
	}

	nextID, err := getNextID()
	if err != nil {
		fmt.Println("Could not get next ID:", err)
		return
	}

	htmlBody := mdToHTML(md)
	html := injectHTML(htmlTemplate, htmlBody, []byte(title))

	outPath := filepath.Join(htmlDir, fmt.Sprintf("%d.html", nextID))
	f, err := os.Create(outPath)
	if err != nil {
		fmt.Println("Could not create file:", err)
		return
	}
	defer f.Close()

	_, err = f.Write(html)
	if err != nil {
		fmt.Println("Could not write to file:", err)
		return
	}

	newItem := Item{
		ID:          nextID,
		Title:       title,
		UpdatedTime: time.Now().Format(time.RFC3339),
	}

	jsonFilePath := os.Getenv("HOME_DIR")
	jsonFilePath += "/blogs.json"
	err = updateJSONFile(jsonFilePath, newItem)
	if err != nil {
		fmt.Println("Could not update JSON file:", err)
		return
	}

	err = updateEnvFile(nextID)
	if err != nil {
		fmt.Println("Could not update env file:", err)
		return
	}

	fmt.Printf("Blog created successfully with ID: %d\n", nextID)
}
