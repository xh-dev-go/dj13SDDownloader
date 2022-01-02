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
//const FORMAT_PNG = "png"
//const FORMAT_JPEG = "jpeg"

func main() {
	//var width, height, quality int
	var dummyBool bool

	var imageFileName, scriptName, updateScriptName, host, pattern string
	var urlOnly, useStdin, useFile, updateFile, useClipboard, out, dryRun bool
	//var useJpeg, usePng bool
	flag.StringVar(&host, "host", "https://sequence.davidje13.com", "The host of processing site")
	flag.StringVar(&pattern, "naming-pattern", "", fmt.Sprintf("common naming pattern for the file naming:\n script-name={pattern}.%s\n img-file={pattern}.{img-extension}\n update-script-name={pattern}.%s\n", SD_EXTENSION, SD_EXTENSION))
	flag.StringVar(&scriptName, "script-name", "", "script file to be loaded, if leave empty, {naming-pattern} will be effective")
	flag.StringVar(&imageFileName, "img-name", "", "output image file name, if leave empty, {naming-pattern} will be effective")
	flag.BoolVar(&useClipboard, "from-clipboard", false, "load data from clipboard")
	flag.BoolVar(&useFile, "from-file", false, "load data from file")
	flag.BoolVar(&useStdin, "from-stdin", true, "[default] load data from stdin")
	flag.BoolVar(&updateFile, "persist", false, "create a local file storing the processed script")
	flag.StringVar(&updateScriptName, "persist-script-name", "", "script file to be saved, if leave empty, {naming-pattern} will be effective")
	//flag.BoolVar(&usePng, "format-png", false, "output image format as png, if no format-{png|svg} set svg will be used")
	//flag.BoolVar(&useJpeg, "format-jpeg", false, "output image format as jpeg, if no format-{png|svg} set svg will be used")
	flag.BoolVar(&dummyBool, "format-svg", false, "[default]output image format as svg")
	//flag.IntVar(&width, "image-width", 0, "for non svg output, width of image")
	//flag.IntVar(&height, "image-height", 0, "for non svg output, height of image")
	//flag.IntVar(&quality, "image-quality", 90, "for jpeg output only, quality of the image, default 90")
	flag.BoolVar(&out, "output-file", false, "output image to file")
	flag.BoolVar(&urlOnly, "output-url", false, "output url for the file")
	flag.BoolVar(&dummyBool, "output-stdout", false, "[default]output svg to stdout")
	flag.BoolVar(&dryRun, "dry-run", false, "only show the step to execute without actual action")
	flag.Parse()

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

	var storeAs string = FORMAT_SVG
	//if useJpeg {
	//	storeAs = FORMAT_JPEG
	//} else if usePng {
	//	storeAs = FORMAT_PNG
	//} else {
	//	storeAs = FORMAT_SVG
	//}
	if out {
		if imageFileName == "" && pattern == "" {
			panic("either {img-file} or {pattern} should be set")
		} else if imageFileName == "" && pattern != "" {
			switch storeAs {
			case FORMAT_SVG:
				imageFileName = pattern + ".svg"
			//case FORMAT_JPEG:
			//	imageFileName = pattern + ".jpeg"
			//case FORMAT_PNG:
			//	imageFileName = pattern + ".png"
			}
		}

		//if useJpeg || usePng {
		//	if width <= 0 || height <= 0 {
		//		panic(fmt.Sprintf("Image width[%d] or image Height[%d] not valid, should in larger than 0", width, height))
		//	}
		//}
		//
		//if useJpeg {
		//	if quality <= 0 || quality > 100 {
		//		panic(fmt.Sprintf("Image quality[%d] not valid, should in range (0,100]", quality))
		//	}
		//}
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
			//if usePng || useJpeg {
			//	fmt.Printf("Width: %d\n", width)
			//	fmt.Printf("Heigh: %d\n", height)
			//}
			//if useJpeg {
			//	fmt.Printf("Image quality: %d\n", quality)
			//}
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

	urlStr := host + "/render/" + diagramScript + ".svg"
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
			//if usePng || useJpeg {
			//	icon, _ := oksvg.ReadIconStream(bytes.NewBuffer(bodyBytes))
			//	icon.SetTarget(0, 0, float64(width), float64(height))
			//	rgba := image.NewRGBA(image.Rect(0, 0, width, height))
			//	icon.Draw(rasterx.NewDasher(width, height, rasterx.NewScannerGV(width, height, rgba, rgba.Bounds())), 1)
			//
			//	out, err := os.Create(imageFileName)
			//	if err != nil {
			//		panic(err)
			//	}
			//	defer out.Close()
			//
			//	if usePng {
			//		err = png.Encode(out, rgba)
			//		if err != nil {
			//			panic(err)
			//		}
			//	} else {
			//		opt := jpeg.Options{
			//			Quality: quality,
			//		}
			//		err = jpeg.Encode(out, rgba, &opt)
			//		if err != nil {
			//			panic(err)
			//		}
			//	}
			//} else {
			//	err := os.WriteFile(imageFileName, bodyBytes, 0644)
			//	if err != nil {
			//		panic(err)
			//	}
			//}
			err := os.WriteFile(imageFileName, bodyBytes, 0644)
			if err != nil {
				panic(err)
			}
		} else {
			fmt.Println(string(bodyBytes))
		}
	}
}
