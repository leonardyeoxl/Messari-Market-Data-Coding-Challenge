package main

import (
	"context"
	"fmt"
	"io"

	"github.com/leonardyeoxl/messari-market-data-coding-challenge/asset"
	"github.com/leonardyeoxl/messari-market-data-coding-challenge/processor"
)

func main() {
	reader, writer := io.Pipe()
	cmdCtx, cmdDone := context.WithCancel(context.Background())
	scannerStopped := make(chan struct{})

	processorService := processor.NewProcessorService()
	cmd, err := processorService.Initialize(writer)
	if err != nil {
		fmt.Println(err.Error())
	}
	assetService := asset.NewAssetService()

	go processorService.Read(assetService, reader, scannerStopped)
	go processorService.Start(cmd, cmdDone, writer)

	<-cmdCtx.Done()
	<-scannerStopped
}
