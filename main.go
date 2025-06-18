package main

import (
	"bytes"
	"context"
	"fmt"
	"image/jpeg"
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"

	"github.com/sampila/uniai-client/pkg/uniai"
	"github.com/unidoc/unipdf/v4/common/license"
	"github.com/unidoc/unipdf/v4/model"
	"github.com/unidoc/unipdf/v4/render"
)

func init() {
	err := license.SetMeteredKey(os.Getenv("UNIDOC_LICENSE_API_KEY_DEV"))
	if err != nil {
		panic(err)
	}
}

func main() {
	if len(os.Args) < 4 {
		println("Usage: go run main.go <path_to_pdf> <output_directory> <prompt>")
		return
	}

	err := godotenv.Load() // by default loads ".env"
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	pdfPath := os.Args[1]
	outputDir := os.Args[2]
	prompt := os.Args[3]

	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		println("PDF file does not exist:", pdfPath)

		return
	}

	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		// Create the output directory if it does not exist
		err := os.MkdirAll(outputDir, 0755)
		if err != nil {
			println("Failed to create output directory:", err.Error())

			return
		}
	}

	pdfFile, err := os.ReadFile(pdfPath)
	if err != nil {
		println("Failed to open PDF file:", err.Error())

		return
	}

	pdfReader, err := model.NewPdfReader(bytes.NewReader(pdfFile))
	if err != nil {
		println("Failed to open PDF file:", err.Error())

		return
	}

	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		println("Failed to get number of pages:", err.Error())

		return
	}

	type renderedPage struct {
		pageNum  int
		filePath string
	}

	renderedPages := make([]renderedPage, numPages)

	var wg sync.WaitGroup

	for i := 1; i <= numPages; i++ {
		wg.Add(1)

		rf := bytes.NewReader(pdfFile)
		newReader, err := model.NewPdfReader(rf)
		if err != nil {
			println("Failed to create new PDF reader for page", i, ":", err.Error())

			wg.Done()
			continue
		}

		go func(pageNum int, pdfReader *model.PdfReader) {
			defer wg.Done()

			page, err := pdfReader.GetPage(i)
			if err != nil {
				println("Failed to get page", i, ":", err.Error())

				return
			}

			device := render.NewImageDevice()
			device.OutputWidth = 1240 // Set the output width for rendering.

			image, err := device.Render(page)
			if err != nil {
				println("Failed to render page", i, ":", err.Error())
				return
			}

			// Save the rendered image to the output directory.
			outputFilePath := outputDir + fmt.Sprintf("/page_%d.png", i)
			outputFile, err := os.Create(outputFilePath)
			if err != nil {
				println("Failed to create output file for page", i, ":", err.Error())
				return
			}

			err = jpeg.Encode(outputFile, image, &jpeg.Options{Quality: 90})
			if err != nil {
				println("Failed to encode image for page", i, ":", err.Error())
				return
			}

			err = outputFile.Close()
			if err != nil {
				println("Failed to close output file for page", i, ":", err.Error())
				return
			}

			renderedPages[i-1] = renderedPage{
				pageNum:  i,
				filePath: outputFilePath,
			}

			// Here you can process the page as needed.
			// For example, you could extract text or images from the page.
			println("Successfully processed page", i)
		}(i, newReader) // Use i - 1 for zero-based index in goroutine
	}
	wg.Wait()

	uniaiClient, err := uniai.NewClient(os.Getenv("API_BASEURL"), nil, os.Getenv("API_AUTH"))
	if err != nil {
		println("Failed to initialize UniAI client:", err.Error())
		return
	}

	// After processing render all pages, request to uniai
	for _, page := range renderedPages {
		println("Rendered page", page.pageNum, "saved to", page.filePath)
		fb, err := os.ReadFile(page.filePath)
		if err != nil {
			println("Failed to read file for page", page.pageNum, ":", err.Error())
			continue
		}

		requestGen := uniai.GenerateRequest{
			Model:   uniai.ModelDefault,
			Prompt:  prompt,
			Images:  []uniai.ImageData{fb},
			System:  "If user mentioned to process with 'high precision', it means prioritize to OCR the image file from request",
			Options: uniai.DefaultOptions,
		}

		println("User prompt:", requestGen.Prompt)
		println("System prompt:", requestGen.System)
		println("Response:")

		ctx := context.Background()
		funcResp := func(resp uniai.GenerateResponse) error {
			// Handle the response from UniAI.
			// For example, you could print the response or save it to a file.
			fmt.Print(resp.Response)

			return nil
		}

		err = uniaiClient.Generate(ctx, &requestGen, funcResp)
		if err != nil {
			println("Failed to generate response for page", page.pageNum, ":", err.Error())
			continue
		}
		fmt.Println()
	}

	// Example: Print a message indicating that the program has started.
	println("UniDoc library initialized successfully. You can now use its features.")
}
