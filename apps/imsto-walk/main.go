package main

import (
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/nfnt/resize"
)

func processImage(inputPath, outputPath string, size uint) error {
	file, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return err
	}

	resizedImg := resize.Resize(size, 0, img, resize.Lanczos3)

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	switch strings.ToLower(filepath.Ext(outputPath)) {
	case ".jpg", ".jpeg":
		err = jpeg.Encode(outputFile, resizedImg, nil)
	case ".png":
		err = png.Encode(outputFile, resizedImg)
	default:
		return fmt.Errorf("Unsupported image format")
	}

	if err != nil {
		return err
	}

	return nil
}

func walkDir(srcDir, destDir string, size uint) error {
	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".jpg" || ext == ".jpeg" || ext == ".png" {
				destPath := filepath.Join(destDir, filepath.Base(path))
				err := processImage(path, destPath, size)
				if err != nil {
					log.Printf("Error processing %s: %v\n", path, err)
				} else {
					fmt.Printf("Processed: %s\n", path)
				}
			}
		}

		return nil
	})

	return err
}

func main() {
	srcDir := flag.String("input", "", "Source directory containing images")
	destDir := flag.String("output", "", "Output directory for resized images")
	size := flag.Uint("size", 1024, "Target size to resize the images")

	flag.Parse()

	if *srcDir == "" || *destDir == "" {
		fmt.Println("Please provide both input and output directories")
		return
	}

	err := walkDir(*srcDir, *destDir, *size)
	if err != nil {
		log.Fatal(err)
	}
}
