package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/leonardyeoxl/messari-market-data-coding-challenge/asset"
	"github.com/leonardyeoxl/messari-market-data-coding-challenge/processor"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	tradeObjectsStdoutBinaryPath := os.Getenv("TRADE_OBJECTS_STDOUT_BINARY_PATH")
	outputPath := os.Getenv("OUTPUT_PATH")

	assetProcessedChannel := make(chan []asset.AssetProcessedResult)
	reader, writer := io.Pipe()
	cmdCtx, cmdDone := context.WithCancel(context.Background())
	scannerStopped := make(chan struct{})

	processorService := processor.NewProcessorService()
	cmd, err := processorService.Initialize(tradeObjectsStdoutBinaryPath, writer)
	if err != nil {
		fmt.Println(err.Error())
	}
	assetService := asset.NewAssetService()

	go processorService.Read(assetService, reader, scannerStopped, assetProcessedChannel)
	go processorService.Start(cmd, cmdDone, writer)

	assetProcessedResults := <-assetProcessedChannel
	processorService.WriteOutput(outputPath, assetProcessedResults)

	<-cmdCtx.Done()
	<-scannerStopped
}
