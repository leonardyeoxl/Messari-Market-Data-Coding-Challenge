package main

import (
	"context"
	"io"
	"os"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"

	"github.com/leonardyeoxl/messari-market-data-coding-challenge/asset"
	"github.com/leonardyeoxl/messari-market-data-coding-challenge/processor"
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetReportCaller(true)
}

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
		log.Fatal(err.Error())
	}
	assetService := asset.NewAssetService()

	go processorService.Read(assetService, reader, scannerStopped, assetProcessedChannel)
	go processorService.Start(cmd, cmdDone, writer)

	assetProcessedResults := <-assetProcessedChannel
	processorService.WriteOutput(outputPath, assetProcessedResults)

	<-cmdCtx.Done()
	<-scannerStopped
}
