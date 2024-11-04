package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"io"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
)

const header = `<!DOCTYPE html>
<html>
<head>
<meta http-equiv="content-type" content="text/html; charset=utf-8">
<title>Markdown Preview Tool</title>
</head>`

const footer = `</body>
			    </html>`

const defaultTemplate = `<!DOCTYPE html>
<html>
<head>
<meta http-equiv="content-type" content="text/html; charset=utf-8">
<title>{{ .Title }}</title>
</head>
<body>
{{ .Body }}
</body>
</html>`

type content struct {
	Title string
	Body  template.HTML
}

func main() {

	// Get the filename from the flags
	filename := flag.String("file", "", "Markdown file to Preview")
	skipPreview := flag.Bool("s", false, "Skip auto-preview")
	tfName := flag.String("t", "", "Template to preview.")
	flag.Parse()

	if *filename == "" {
		flag.Usage()
		os.Exit(1)
	}

	if err := run(*filename, os.Stdout, *skipPreview, *tfName); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

func run(filename string, out io.Writer, skipReview bool, tfName string) error {
	input, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	htmlData, err := parseContent(input, tfName)
	if err != nil {
		return err
	}

	temp, err := os.CreateTemp("", "mdp*.html")

	if err != nil {
		return err
	}

	if err = temp.Close(); err != nil {
		return err
	}

	outName := temp.Name()

	fmt.Println(outName)

	if err = saveHTML(outName, htmlData); err != nil {
		return err
	}

	if !skipReview {
		return nil
	}

	defer os.Remove(outName)

	return preview(outName)

}

func parseContent(input []byte, tFname string) ([]byte, error) {
	output := blackfriday.Run(input)
	body := bluemonday.UGCPolicy().SanitizeBytes(output)

	t, err := template.New("mdp").Parse(defaultTemplate)
	if err != nil {
		return nil, err
	}

	if tFname != "" {
		t, err = template.ParseFiles(tFname)
		if err != nil {
			return nil, err
		}
	}

	c := content{
		Title: "Markdown Preview Tool",
		Body:  template.HTML(body),
	}

	var buffer bytes.Buffer
	if err := t.Execute(&buffer, c); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func saveHTML(outFName string, data []byte) error {
	return os.WriteFile(outFName, data, 0644)
}

func preview(fname string) error {
	cName := ""
	cParams := []string{}

	switch runtime.GOOS {
	case "linux":
		cName = "xdg-open"
	case "windows":
		cName = "cmd.exe"
		cParams = []string{"/C", "start"}
	case "darwin":
		cName = "open"
	default:
		return fmt.Errorf("OS not supported")
	}
	cParams = append(cParams, fname)
	fmt.Println(cName)
	cPath, err := exec.LookPath(cName)

	if err != nil {
		return err
	}

	err = exec.Command(cPath, cParams...).Run()

	time.Sleep(2 * time.Second)

	return err
}
