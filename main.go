package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/atotto/clipboard"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func ReadFromFile(file string) string {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	return string(bytes)
}
func ReadFromStd() string {
	reader := bufio.NewReader(os.Stdin)
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		panic(err)
	}
	return string(bytes)
}
func ReadFromClipboard() string {
	data, err := clipboard.ReadAll()
	if err != nil {
		println(data)
		panic(err)
	}
	return data
}

const SD_EXTENSION = "sddsl"
const FORMAT_SVG = "svg"

const Version = "1.0.1 - 2022-01-21"

func main() {
	var dummyBool bool

	var imageFileName, scriptName, updateScriptName, host, pattern string
	var pngFileName string
	var urlOnly, useStdin, useFile, updateFile, useClipboard, out, dryRun, showVersion, withPng bool
	var pngScale int
	flag.BoolVar(&showVersion, "version", false, "Show the software version")
	flag.StringVar(&host, "host", "https://sequence.davidje13.com", "The host of processing site")
	flag.StringVar(&pattern, "naming-pattern", "", fmt.Sprintf("common naming pattern for the file naming:\n script-name={pattern}.%s\n img-file={pattern}.{img-extension}\n update-script-name={pattern}.%s\n", SD_EXTENSION, SD_EXTENSION))
	flag.StringVar(&scriptName, "script-name", "", "script file to be loaded, if leave empty, {naming-pattern} will be effective")
	flag.StringVar(&imageFileName, "img-name", "", "output image file name, if leave empty, {naming-pattern} will be effective")
	flag.BoolVar(&useClipboard, "from-clipboard", false, "load data from clipboard")
	flag.BoolVar(&useFile, "from-file", false, "load data from file")
	flag.BoolVar(&useStdin, "from-stdin", true, "[default] load data from stdin")
	flag.BoolVar(&updateFile, "persist", false, "create a local file storing the processed script")
	flag.StringVar(&updateScriptName, "persist-script-name", "", "script file to be saved, if leave empty, {naming-pattern} will be effective")
	flag.BoolVar(&dummyBool, "format-svg", false, "[default]output image format as svg")
	flag.BoolVar(&out, "output-file", false, "output image to file")
	flag.BoolVar(&urlOnly, "output-url", false, "output url for the file")
	flag.BoolVar(&dummyBool, "output-stdout", false, "[default]output svg to stdout")
	flag.BoolVar(&dryRun, "dry-run", false, "only show the step to execute without actual action")
	flag.BoolVar(&withPng, "export-png", false, "also export as png")
	flag.IntVar(&pngScale, "png-scale", 3, "scale to {png-scale}X of origin svg")
	flag.Parse()

	if showVersion {
		fmt.Println("version: "+Version)
		os.Exit(0)
	}

	if useFile {
		if scriptName == "" && pattern == "" {
			panic("either {script-name} or {pattern} should be set")
		} else if scriptName == "" && pattern != "" {
			scriptName = pattern + "." + SD_EXTENSION
		}
	}
	if updateFile {
		if updateScriptName == "" && pattern == "" {
			panic("either {update-script-name} or {pattern} should be set")
		} else if updateScriptName == "" && pattern != "" {
			updateScriptName = pattern + "." + SD_EXTENSION
		}
	}

	var storeAs = FORMAT_SVG
	svgSuffix := ".svg"
	pngSuffix := ".png"
	if out {
		if imageFileName == "" && pattern == "" {
			panic("either {img-file} or {pattern} should be set")
		} else if imageFileName == "" && pattern != "" {
			switch storeAs {
			case FORMAT_SVG:
				imageFileName = pattern + svgSuffix
			}
		}

		if withPng {
			if strings.HasSuffix(imageFileName,svgSuffix){
				pngFileName = imageFileName[0: len(imageFileName)-len(svgSuffix)]+pngSuffix
			} else {
				pngFileName = imageFileName+pngSuffix
			}
		}
	}

	if dryRun {
		fmt.Println("=========== Dry run ===========")
		fmt.Println("Process host: " + host)
		if useClipboard {
			fmt.Println("Data from: Clipboard")
		} else if useFile {
			fmt.Println("Data from: File")
			fmt.Println("Load data from: " + scriptName)
		} else {
			fmt.Println("Data from: Stdin")
		}
		if updateFile {
			fmt.Println("Update file: " + updateScriptName)
		}
		if out {
			fmt.Println("Output image: " + imageFileName)
			if withPng {
				fmt.Println("output with png: "+ imageFileName)
			}
		}

		os.Exit(0)
	}

	var diagramScript string
	if useClipboard {
		diagramScript = ReadFromClipboard()
	} else if useFile {
		diagramScript = ReadFromFile(scriptName)
	} else {
		diagramScript = ReadFromStd()
	}
	originDiagramScript := diagramScript

	tempReplacing := "!@#$)(*"
	diagramScript = strings.ReplaceAll(diagramScript, "/", tempReplacing)
	var re = regexp.MustCompile("(\n+)")
	diagramScript = re.ReplaceAllString(diagramScript, `/`)
	diagramScript = strings.ReplaceAll(diagramScript, tempReplacing, "/")
	diagramScript = url.QueryEscape(diagramScript)
	diagramScript = strings.ReplaceAll(diagramScript, "%2F", "/")
	diagramScript = strings.ReplaceAll(diagramScript, "+", "%20")

	urlStr := host + "/render/" + diagramScript + svgSuffix
	if urlOnly {
		fmt.Println(urlStr)
		os.Exit(0)
	}
	response, err := http.Get(urlStr)
	if err != nil {
		panic(err)
	} else if response.StatusCode != 200 {
		panic("error downloading image")
	} else if bodyBytes, err := io.ReadAll(response.Body); err != nil {
		panic(err)
	} else {
		if updateFile {
			err := os.WriteFile(updateScriptName, []byte(originDiagramScript), 0644)
			if err != nil {
				panic(err)
			}
		}
		if out {
			err := os.WriteFile(imageFileName, bodyBytes, 0644)
			if err != nil {
				panic(err)
			}

			cmd := exec.Command("svgexport", imageFileName, pngFileName, fmt.Sprintf("%dx", pngScale))
			cmd.Run()
		} else {
			fmt.Println(string(bodyBytes))
		}
	}
}
