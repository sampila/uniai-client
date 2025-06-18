package cli

import (
	"errors"
	"fmt"
	"image/jpeg"
	"os"

	"github.com/unidoc/unipdf/v4/model"
	"github.com/unidoc/unipdf/v4/render"
)

func RenderPdfPage(pageNumber int, page *model.PdfPage, outputDir string) (string, error) {
	if page == nil {
		return "", errors.New("page is nil")
	}

	device := render.NewImageDevice()
	device.OutputWidth = 1400

	img, err := device.Render(page)
	if err != nil {
		return "", err
	}

	outputFilePath := outputDir + fmt.Sprintf("/page_%d.jpg", pageNumber)

	f, err := os.Create(outputFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	err = jpeg.Encode(f, img, &jpeg.Options{Quality: 90})
	if err != nil {
		return "", fmt.Errorf("failed to encode image: %w", err)
	}

	return outputFilePath, nil
}
