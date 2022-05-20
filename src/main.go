package main

import (
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/gen2brain/go-fitz"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font/gofont/goregular"
)

var (
	pdfsDirPath *string = flag.String(
		"src_dir",
		".",
		"Specify path to the directory where each found PDF file's pages will be converted to images",
	)

	extractionDirPath *string = flag.String(
		"dst_dir",
		"output_images",
		"Specify path to the output directory",
	)

	note *string = flag.String(
		"note",
		"",
		"Set a note that will be added to the bottom of each extracted image",
	)

	fontSize *float64 = flag.Float64(
		"fontsize",
		150.0,
		"Set font size for note",
	)

	addSpace *bool = flag.Bool(
		"add_note_space",
		false,
		"Add bottom space for a note or not",
	)
)

func addText(img *image.RGBA, x, y uint, text string, size float64) error {
	ctx := freetype.NewContext()
	ctx.SetDPI(72)
	newFont, err := truetype.Parse(goregular.TTF)
	if err != nil {
		return err
	}
	ctx.SetFont(newFont)
	ctx.SetFontSize(size)
	ctx.SetClip(img.Bounds())
	ctx.SetSrc(image.Black)
	ctx.SetDst(img)

	pt := freetype.Pt(int(x), int(y)+int(ctx.PointToFixed(size)>>6))
	_, err = ctx.DrawString(text, pt)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	flag.Parse()

	err := os.MkdirAll(*extractionDirPath, os.ModePerm)
	if err != nil {
		fmt.Printf("Could not create output directory: %s\n", err)
		os.Exit(1)
	}

	entries, err := os.ReadDir(*pdfsDirPath)
	if err != nil {
		fmt.Printf("Could not read specified directory fully: %s\n", err)
		if len(entries) == 0 {
			os.Exit(1)
		}
	}

	for count, entry := range entries {
		entryName := entry.Name()
		entryPath := filepath.Join(*pdfsDirPath, entryName)
		saveDirPath := filepath.Join(*extractionDirPath, entryName)

		if !strings.HasSuffix(entryName, ".pdf") || entry.IsDir() {
			fmt.Printf("[%d] Skipping %s: not a PDF file\n", count, entryName)
			continue
		}
		fmt.Printf("[%d] Working with %s...\n", count, entryName)

		pdf, err := fitz.New(entryPath)
		if err != nil {
			fmt.Printf("[%d] Could not read %s: %s", count, entryName, err)
			continue
		}

		err = os.MkdirAll(saveDirPath, os.ModePerm)
		if err != nil {
			fmt.Printf("[%d] Could not make extraction directory for %s: %s\n", count, entryName, err)
			continue
		}

		for i := 0; i < pdf.NumPage(); i++ {
			srcPDFImage, err := pdf.Image(i)
			if err != nil {
				fmt.Printf("[%d] Could not extract page as image from %s, page %d: %s\n", count, entryName, i, err)
				continue
			}

			outputImageFile, err := os.Create(
				filepath.Join(saveDirPath,
					strings.TrimSuffix(entryName, ".pdf")+fmt.Sprintf("_%d", i)+".png"))
			if err != nil {
				fmt.Printf("[%d] Could not create image file for %s page %d: %s\n", count, entryName, i, err)
				continue
			}

			switch *note != "" {
			case true:
				var newImage *image.RGBA
				if *addSpace {
					extendedImageDimensions := image.Rectangle{
						image.Pt(0, 0),
						image.Pt(srcPDFImage.Bounds().Dx(), srcPDFImage.Bounds().Dy()+int(*fontSize)+int(*fontSize/2)),
					}
					newImage = image.NewRGBA(extendedImageDimensions)
					draw.Draw(newImage, newImage.Bounds(), image.White, image.Pt(0, 0), draw.Src)
				} else {
					newImage = image.NewRGBA(srcPDFImage.Bounds())
				}
				draw.Draw(newImage, srcPDFImage.Bounds(), srcPDFImage, image.Pt(0, 0), draw.Src)

				err = addText(newImage, 0, uint(srcPDFImage.Bounds().Dy()-int(*fontSize+*fontSize/4)), *note, *fontSize)
				if err != nil {
					fmt.Printf("[%d] Could not add text to %s, page %d: %s. Saving without additions...\n", count, entryName, i, err)
					png.Encode(outputImageFile, srcPDFImage)
					continue
				}

				png.Encode(outputImageFile, newImage)
			case false:
				png.Encode(outputImageFile, srcPDFImage)
			}

			outputImageFile.Close()
		}
	}
}
