package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/nasjp/pinata-example/pinata"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	client := pinata.New(
		pinata.PinataAPIKey,
		pinata.PinataAPISecret,
	)

	if err := client.Healthcheck(); err != nil {
		return err
	}

	const targetDir = "data/images"

	files, err := os.ReadDir(targetDir)
	if err != nil {
		return err
	}

	imageContents := make([][]byte, 0)
	images := make([]string, 0)

	for _, entry := range files {
		info, err := entry.Info()
		if err != nil {
			return err
		}

		if info.IsDir() {
			continue
		}

		content, err := os.ReadFile(filepath.Join(targetDir, info.Name()))
		if err != nil {
			return err
		}

		imageContents = append(imageContents, content)
		images = append(images, info.Name())
	}

	pinataMetadata := &pinata.PinataMetadata{
		Name: "nft",
		KeyValues: map[string]string{
			"exampleKey": "exampleValue",
		},
	}

	imageResult, err := client.PinDir(imageContents, images, "images", pinataMetadata)
	if err != nil {
		return err
	}

	fmt.Printf("image cid: (%s)\n", imageResult.IPFSHash)

	metadataContents := make([][]byte, 0)
	metadatas := make([]string, 0)

	for _, name := range images {

		metadata := &Metadata{
			Image: "ipfs://" + path.Join(imageResult.IPFSHash, name),
		}

		jsonBuf := bytes.NewBuffer(nil)
		if err := json.NewEncoder(jsonBuf).Encode(metadata); err != nil {
			return err
		}

		metadataContents = append(metadataContents, jsonBuf.Bytes())
		metadatas = append(metadatas, filepath.Base(name)+".json")
	}

	metadataResult, err := client.PinDir(metadataContents, metadatas, "metadata", pinataMetadata)
	if err != nil {
		return err
	}

	fmt.Printf("metadata cid: (%s)\n", metadataResult.IPFSHash)

	return nil
}

type Metadata struct {
	Image string `json:"image"`
}
