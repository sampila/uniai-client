package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/cobra"

	"github.com/unidoc/unipdf/v4/model"

	"github.com/sampila/uniai-client/internal/cli"
	"github.com/sampila/uniai-client/pkg/uniai"
)

var (
	filePath      string
	outputDir     string
	prompt        string
	pageRange     string // e.g., "1-3" for pages 1 to 3, "1,2,4" for specific pages
	isParallel    bool   // Flag to indicate if processing should be parallelized
	writeResponse bool   // Flag to indicate if the response should be written to a file
)

var uniaiCmd = &cobra.Command{
	Use:   "uniai",
	Short: "UniAI is a CLI client for interacting with UniAI models.",
	Long: `UniAI is a command-line interface (CLI) client designed to interact with UniAI models,
providing functionalities such as pdf to text generation, document QA, and make structured data.`,
	Run: func(cmd *cobra.Command, args []string) {
		if filePath == "" || outputDir == "" || prompt == "" {
			cmd.Help()
			return
		}

		var (
			pageNumbers []int
			err         error
		)
		if pageRange != "" {
			pageNumbers, err = cli.ParsePageRange(pageRange)
			if err != nil {
				println("Invalid page range format:", err.Error())
				return
			}
		}

		// Read the file and process it
		fp, err := os.ReadFile(filePath)
		if err != nil {
			println("Failed to read file:", err.Error())
			return
		}

		pdfReader, err := model.NewPdfReader(bytes.NewReader(fp))
		if err != nil {
			println("Failed to open PDF file:", err.Error())
			return
		}

		numPages, err := pdfReader.GetNumPages()
		if err != nil {
			println("Failed to get number of pages:", err.Error())
			return
		}

		if len(pageNumbers) == 0 {
			// If no specific pages are provided, process all pages
			for i := 1; i <= numPages; i++ {
				pageNumbers = append(pageNumbers, i)
			}
		}

		type renderedPage struct {
			pageNum  int
			filePath string
		}
		renderedPages := make([]renderedPage, numPages)

		var (
			wg  sync.WaitGroup
			sem = make(chan struct{}, 3) // Semaphore to limit concurrency
		)

		base := filepath.Base(filePath) // "report 2025.pdf"
		dirName := strings.TrimSuffix(base, filepath.Ext(base))

		outDir := filepath.Join(outputDir, dirName)
		if _, err := os.Stat(outDir); os.IsNotExist(err) {
			err = os.MkdirAll(outDir, 0755)
			if err != nil {
				println("Failed to create output directory:", err.Error())
				return
			}
		}

		for _, pageNum := range pageNumbers {
			if pageNum < 1 || pageNum > numPages {
				println("Page number out of range:", pageNum)
				continue
			}

			if isParallel {
				wg.Add(1)
				sem <- struct{}{} // Acquire a semaphore slot
				go func(pageNum int) {
					defer wg.Done()
					defer func() { <-sem }()

					newReader, err := model.NewPdfReader(bytes.NewReader(fp))
					page, err := newReader.GetPage(pageNum)
					if err != nil {
						println("Failed to get page:", err.Error())
						return
					}

					// Render the page to an image
					output, err := cli.RenderPdfPage(pageNum, page, outDir)
					if err != nil {
						println("Failed to render page:", err.Error())
						return
					}
					renderedPages[pageNum-1] = renderedPage{
						pageNum:  pageNum,
						filePath: output,
					}
					println("Rendered page", pageNum, "to", output)
				}(pageNum)
			} else {
				page, err := pdfReader.GetPage(pageNum)
				if err != nil {
					println("Failed to get page:", err.Error())
					continue
				}

				// Render the page to an image
				output, err := cli.RenderPdfPage(pageNum, page, outputDir)
				if err != nil {
					println("Failed to render page:", err.Error())
					continue
				}
				renderedPages[pageNum-1] = renderedPage{
					pageNum:  pageNum,
					filePath: output,
				}
				println("Rendered page", pageNum, "to", output)
			}
		}
		wg.Wait()

		// Init UniAI client
		uniaiClient, err := uniai.NewClient(os.Getenv("API_BASEURL"), nil, os.Getenv("API_AUTH"))
		if err != nil {
			println("Failed to initialize UniAI client:", err.Error())
			return
		}

		for _, page := range renderedPages {
			println("Rendered page", page.pageNum, "saved to", page.filePath)
			fb, err := os.ReadFile(page.filePath)
			if err != nil {
				println("Failed to read file for page", page.pageNum, ":", err.Error())
				continue
			}

			if writeResponse {
				var (
					respDir          string
					responseFilePath string
					rf               *os.File
				)
				// write response to a in directory response
				respDir = filepath.Join(outDir, "response")
				if _, err := os.Stat(respDir); os.IsNotExist(err) {
					err = os.MkdirAll(respDir, 0755)
					if err != nil {
						println("Failed to create response directory:", err.Error())
						continue
					}
				}
				responseFilePath = filepath.Join(respDir, fmt.Sprintf("page_%d.txt", page.pageNum))
				rf, err = os.Create(responseFilePath)
				if err != nil {
					println("Failed to create response file for page", page.pageNum, ":", err.Error())
					continue
				}
				defer rf.Close()

				os.Stderr = rf // Redirect stderr to the response file
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
			if writeResponse {
				println("Response written to file")
			}

			ctx := context.Background()
			funcResp := func(resp uniai.GenerateResponse) error {
				// Handle the response from UniAI.
				// For example, you could print the response or save it to a file.
				fmt.Fprint(os.Stderr, resp.Response)
				if resp.Done {
					fmt.Fprintln(os.Stderr)
					resp.Summary()
				}

				return nil
			}

			err = uniaiClient.Generate(ctx, &requestGen, funcResp)
			if err != nil {
				println("Failed to generate response for page", page.pageNum, ":", err.Error())
				continue
			}
			fmt.Println()
		}
	},
}

func init() {
	uniaiCmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the input file (PDF or text)")
	uniaiCmd.Flags().StringVarP(&outputDir, "output", "o", "./output", "Directory to save the output files")
	uniaiCmd.Flags().StringVarP(&prompt, "prompt", "m", "", "Prompt for the model (required for some commands)")
	uniaiCmd.Flags().StringVarP(&pageRange, "pages", "r", "", "Page range to process (e.g., '1-3' for pages 1 to 3, '1,2,4' for specific pages)")
	uniaiCmd.Flags().BoolVarP(&isParallel, "parallel", "p", false, "Enable parallel processing of pages (if applicable)")
	uniaiCmd.Flags().BoolVarP(&writeResponse, "write-response", "w", false, "Write the response to a file (if applicable)")

	uniaiCmd.MarkFlagRequired("file")
	uniaiCmd.MarkFlagRequired("prompt")
	uniaiCmd.MarkFlagRequired("output")

	rootCmd.AddCommand(uniaiCmd)
}
