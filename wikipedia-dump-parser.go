// Modified version of - https://github.com/dps/go-xml-parse

package main

import (
	"bufio"
	"encoding/xml"
	"flag"
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"
)

var inputFile = flag.String("infile", "", "Input file path")
var outDir = flag.String("outdir", "", "Output directory")

// Here is an example article from the Wikipedia XML dump
//
// <page>
// 	<title>Apollo 11</title>
//      <redirect title="Foo bar" />
// 	...
// 	<revision>
// 	...
// 	  <text xml:space="preserve">
// 	  {{Infobox Space mission
// 	  |mission_name=&lt;!--See above--&gt;
// 	  |insignia=Apollo_11_insignia.png
// 	...
// 	  </text>
// 	</revision>
// </page>
//
// Note how the tags on the fields of Page and Redirect below
// describe the XML schema structure.

type Redirect struct {
	Title string `xml:"title,attr"`
}

type Page struct {
	Title string   `xml:"title"`
	Redir Redirect `xml:"redirect"`
	Text  string   `xml:"revision>text"`
}

func CanonicalizeTitle(title string) string {
	can := strings.ToLower(title)
	can = strings.Replace(can, " ", "_", -1)
	can = url.QueryEscape(can)
	return can
}

func WritePage(title string, text string, targetFile string) {
	outFile, err := os.Create(targetFile)
	if err == nil {
		writer := bufio.NewWriter(outFile)
		defer outFile.Close()
		writer.WriteString(title)
		writer.WriteString("\n")
		writer.WriteString("---------------\n")
		writer.WriteString(text)
		writer.Flush()
	}
}

func main() {
	flag.Parse()

	xmlFile, err := os.Open(*inputFile)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer xmlFile.Close()

	if *outDir == "" {
		fmt.Println("Outdir is not specified")
		return
	}
	err = os.MkdirAll(path.Join(*outDir, "media-wiki-dump-splitted"), 0777)
	if err != nil {
		fmt.Println("Error creating out directory:", err)
		return
	}

	decoder := xml.NewDecoder(xmlFile)
	total := 0
	var inElement string
	for {
		// Read tokens from the XML document in a stream.
		t, _ := decoder.Token()
		if t == nil {
			break
		}
		// Inspect the type of the token just read.
		switch se := t.(type) {
		case xml.StartElement:
			// If we just read a StartElement token
			inElement = se.Name.Local
			// ...and its name is "page"
			if inElement == "page" {
				var p Page
				// decode a whole chunk of following XML into the
				// variable p which is a Page (se above)
				decoder.DecodeElement(&p, &se)

				// Do some stuff with the page.
				p.Title = CanonicalizeTitle(p.Title)
				if p.Redir.Title == "" {
					targetFile := path.Join(*outDir, "media-wiki-dump-splitted", fmt.Sprintf("%d.txt", total))
					go WritePage(p.Title, p.Text, targetFile)
					total++
				}
			}
		default:
		}

	}

	fmt.Printf("Total articles: %d \n", total)
}
