package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"encoding/json"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/joho/godotenv"
)

type Item struct {
	Title       string `json:"title"`
	UpdatedTime string `json:"updated_time"`
}

func updateJSONFile(filePath string, newTitle string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	var items []Item
	if err := json.Unmarshal(data, &items); err != nil {
		return fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	newItem := Item{
		Title:       newTitle,
		UpdatedTime: time.Now().Format(time.RFC3339), // Format the current time
	}

	items = append(items, newItem)
	updatedData, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %v", err)
	}
	if err := os.WriteFile(filePath, updatedData, 0644); err != nil {
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
	if htmlDir == "" || templatePath == "" {
		fmt.Println("Missing HTML_DIR or TEMPLATE_PATH in env")
		return
	}

	fmt.Println("Enter path of .md file:")
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		fmt.Println("Couldn't read from stdin")
		return
	}
	mdPath := scanner.Text()
	if filepath.Ext(mdPath) != ".md" {
		fmt.Println("File must have .md extension")
		return
	}

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

	filename := filepath.Base(mdPath)
	name := filename[:len(filename)-3]

	htmlBody := mdToHTML(md)
	html := injectHTML(htmlTemplate, htmlBody, []byte(name))

	outDir := filepath.Join(htmlDir, name)
	err = os.MkdirAll(outDir, 0755)
	if err != nil {
		fmt.Println("Could not create directory:", err)
		return
	}

	outPath := filepath.Join(outDir, name+".html")
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
	jsonFilePath := os.Getenv("HOME_DIR/blogs.json")
	updateJSONFile(jsonFilePath, name)
}
