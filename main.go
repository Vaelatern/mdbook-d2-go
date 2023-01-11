package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/jsternberg/markdownfmt/markdown"
	blackfriday "gopkg.in/russross/blackfriday.v2"

	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2layouts/d2elklayout"
	"oss.terrastruct.com/d2/d2lib"
	"oss.terrastruct.com/d2/d2renderers/d2svg"
	"oss.terrastruct.com/d2/d2themes/d2themescatalog"
	"oss.terrastruct.com/d2/lib/textmeasure"
)

type Config struct {
	Config struct {
		Book struct {
			Authors      []string `json:"authors"`
			Language     string   `json:"language"`
			Multilingual bool     `json:"multilingual"`
			Src          string   `json:"src"`
			Title        string   `json:"title"`
		} `json:"book"`
		Output       interface{} `json:"output"`
		Preprocessor struct {
			// Config I'm expecting goes here
			// Unmarshal will leave what we don't have as null
			D2_Go struct {
				Layout string `json:"layout"`
			} `json:"d2-go"`
		} `json:"preprocessor"`
	} `json:"config"`
	MdbookVersion string `json:"mdbook_version"`
	Renderer      string `json:"renderer"`
	Root          string `json:"root"`
}

type BookItem struct {
	Content     string        `json:"content"`
	Name        string        `json:"name"`
	Number      []int         `json:"number"`
	ParentNames []interface{} `json:"parent_names"`
	Path        *string       `json:"path"`
	SourcePath  *string       `json:"source_path"`
	SubItems    []Chapter     `json:"sub_items"`
}

type Chapter struct {
	Chapter BookItem `json:"Chapter"`
}

type Book struct {
	NonExhaustive interface{} `json:"__non_exhaustive"`
	Sections      []Chapter   `json:"sections"`
}

func generateSvgFromD2(graph string) ([]byte, error) {
	ruler, err := textmeasure.NewRuler()
	defaultLayout := func(ctx context.Context, g *d2graph.Graph) error {
		return d2elklayout.Layout(ctx, g, nil)
	}
	if err != nil {
		return nil, err
	}

	diagram, _, err := d2lib.Compile(context.Background(), graph, &d2lib.CompileOptions{
		Layout:  defaultLayout,
		Ruler:   ruler,
		ThemeID: d2themescatalog.GrapeSoda.ID,
	})
	if err != nil {
		return nil, err
	}

	out, err := d2svg.Render(diagram, &d2svg.RenderOpts{
		Pad: d2svg.DEFAULT_PADDING,
	})
	if err != nil {
		return nil, err
	}

	outWithoutConfusingMarkdown := bytes.ReplaceAll(out, []byte("\n\n"), []byte("\n"))

	return outWithoutConfusingMarkdown, nil
}

// needed because markdown only respects the html tags it knows, maybe not <svg>
func wrapSvgInDiv(in []byte) []byte {
	start := "<div>"
	end := "</div>"
	totalLen := len(start) + len(in) + len(end)
	finalSlice := make([]byte, totalLen)
	position := copy(finalSlice[0:], []byte(start))
	position += copy(finalSlice[position:], in)
	position += copy(finalSlice[position:], []byte(end))
	return finalSlice
}

func parseStdin() (Config, Book, error) {
	var toProcess []interface{} = make([]interface{}, 2)

	var config Config
	var book Book

	if err := json.NewDecoder(os.Stdin).Decode(&toProcess); err != nil {
		log.Println("JSON decoder failed")
		return Config{}, Book{}, err
	}

	configStr, err := json.Marshal(toProcess[0])
	if err != nil {
		log.Println("JSON re-encoder the first failed")
		return Config{}, Book{}, err
	}
	if err := json.Unmarshal(configStr, &config); err != nil {
		log.Println("Unmarshalling config failed")
		return Config{}, Book{}, err
	}

	bookStr, err := json.Marshal(toProcess[1])
	if err != nil {
		log.Println("JSON re-encoder the second failed")
		return Config{}, Book{}, err
	}
	if err := json.Unmarshal(bookStr, &book); err != nil {
		log.Println("Unmarshalling book failed")
		return Config{}, Book{}, err
	}

	return config, book, nil
}

func rewriteD2(whoami *blackfriday.Node, entering bool) blackfriday.WalkStatus {
	if whoami.Type != blackfriday.CodeBlock {
		return blackfriday.GoToNext
	}

	if !bytes.Equal(whoami.CodeBlockData.Info, []byte("d2")) {
		return blackfriday.GoToNext
	}

	newSvg, err := generateSvgFromD2(string(whoami.Literal))
	if err != nil {
		newNode := new(blackfriday.Node)
		newNode.Type = blackfriday.BlockQuote
		newNode.Literal = []byte(fmt.Sprintf("Error parsing the below code block into d2: %s", err))
		whoami.InsertBefore(newNode)
		log.Println("Error parsing code block into d2: %s", err)
		return blackfriday.GoToNext
	}
	newText := wrapSvgInDiv(newSvg)
	newNode := new(blackfriday.Node)
	newNode.Type = blackfriday.HTMLBlock
	newNode.Literal = newText
	whoami.InsertBefore(newNode)
	whoami.Unlink()
	log.Println("Found and processed a d2 graph")
	return blackfriday.GoToNext
}

func fromMarkdownThroughD2ToMarkdown(content []byte) ([]byte, error) {
	renderToMarkdown := markdown.NewRenderer(nil)
	opts := []blackfriday.Option{
		blackfriday.WithRenderer(renderToMarkdown),
		blackfriday.WithExtensions(blackfriday.NoIntraEmphasis |
			blackfriday.Tables |
			blackfriday.FencedCode |
			blackfriday.Autolink |
			blackfriday.Strikethrough |
			blackfriday.SpaceHeadings |
			blackfriday.NoEmptyLineBeforeBlock)}
	parser := blackfriday.New(opts...)
	root := parser.Parse(content)

	root.Walk(rewriteD2)

	var buf bytes.Buffer
	renderToMarkdown.RenderHeader(&buf, root)
	root.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		return renderToMarkdown.RenderNode(&buf, node, entering)
	})
	renderToMarkdown.RenderFooter(&buf, root)
	return buf.Bytes(), nil
}

func replaceContent(bookItem *BookItem) error {
	newContent, err := fromMarkdownThroughD2ToMarkdown([]byte(bookItem.Content))
	if err != nil {
		return err
	}
	bookItem.Content = string(newContent)
	for i := range bookItem.SubItems {
		err = replaceContent(&bookItem.SubItems[i].Chapter)
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	log.SetOutput(os.Stderr)
	var usage = func() {
		fmt.Println("This should be used as documented at https://rust-lang.github.io/mdBook/for_developers/preprocessors.html")
		os.Exit(1)
	}
	if len(os.Args[1:]) == 2 {
		if os.Args[1] != "supports" {
			usage()
		}
		if os.Args[2] != "html" {
			log.Fatal("Reader does not support " + os.Args[2])
		}
		os.Exit(0)
	} else if len(os.Args[1:]) != 0 {
		usage()
	}

	config, book, err := parseStdin()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Config: %s", config)

	for i := range book.Sections {
		err = replaceContent(&book.Sections[i].Chapter)
		if err != nil {
			log.Fatal(err)
		}
	}

	outStr, err := json.Marshal(book)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(outStr))
	log.Println("Finished printing d2 processed JSON")
}
