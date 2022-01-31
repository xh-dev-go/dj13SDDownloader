package main

import (
	"bufio"
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"github.com/atotto/clipboard"
	"io"
	"io/ioutil"
	"mime/multipart"
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

//go:embed version
var Version string

var client = &http.Client{}

func uploadFile(url string, file *os.File, outFileName string, scale int) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	if fw, err := w.CreateFormFile("upload", file.Name()); err != nil {
		panic(err)
	} else if _, err := io.Copy(fw, file); err != nil {
		panic(err)
	}

	err := w.Close()
	if err != nil {
		return
	}

	// Now that you have a form, you can submit it to your handler.
	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return
	}
	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("X-SCALE-TO", fmt.Sprintf("%d",scale))

	// Submit the request
	res, err := client.Do(req)
	if err != nil {
		return
	}

	// Check the response
	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("bad status: %s", res.Status)
		panic(err)
	}

	outFile, err := os.Create(outFileName)
	if err != nil {
		return
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, res.Body)
	if err != nil {
		return
	}

	return
}

func main() {
	var dummyBool bool

	var imageFileName, scriptName, updateScriptName, host, pattern string
	var pngFileName, pngWebUrl string
	var urlOnly, useStdin, useFile, updateFile, useClipboard, out, dryRun, showVersion, withPng, pngLocal bool
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
	flag.BoolVar(&pngLocal, "png-local", false, "use local node svgExport rather than png web svgExport service")
	flag.StringVar(&pngWebUrl, "png-web-url", "https://convert-svg-png.mytools.express", "the web svg to png service url")
	flag.IntVar(&pngScale, "png-scale", 3, "scale to {png-scale}X of origin svg")
	flag.Parse()

	if len(os.Args) == 1 {
		flag.Usage()
		os.Exit(0)
	}

	if showVersion {
		fmt.Println("version: " + Version)
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
			if strings.HasSuffix(imageFileName, svgSuffix) {
				pngFileName = imageFileName[0:len(imageFileName)-len(svgSuffix)] + pngSuffix
			} else {
				pngFileName = imageFileName + pngSuffix
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
				fmt.Println("output with png: " + pngFileName)
			}
			if pngLocal {
				fmt.Println("output png with local command svgExport.(Please ensure svgExport is installed. npm install -g svgExport)")
			} else {
				fmt.Printf("output png with web svg to png service. (host: %s)\n", pngWebUrl)
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
	response, err := client.Get(urlStr)
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

			if pngLocal {
				cmd := exec.Command("svgexport", imageFileName, pngFileName, fmt.Sprintf("%dx", pngScale))
				cmd.Run()
			} else {
				fileInput,err := os.Open(imageFileName)
				if err != nil {
					panic(err)
				}
				uploadFile(pngWebUrl, fileInput, pngFileName, pngScale)
			}
		} else {
			fmt.Println(string(bodyBytes))
		}
	}
}
