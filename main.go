package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/leonardyeoxl/messari-market-data-coding-challenge/asset"
)

func main() {
	reader, writer := io.Pipe()
	cmdCtx, cmdDone := context.WithCancel(context.Background())
	scannerStopped := make(chan struct{})

	cache := map[int]asset.AssetProcessTotal{}
	assetService := asset.NewAssetService()

	go func(assetService asset.IAssetService, cache map[int]asset.AssetProcessTotal) {
		defer close(scannerStopped)

		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			line := scanner.Text()
			line = strings.TrimSuffix(line, "\n")
			if line == "" ||
				line == "BEGIN" ||
				line == "END" ||
				strings.Contains(line, "Trade Count") ||
				strings.Contains(line, "Market Count") ||
				strings.Contains(line, "Duration of send operation") {
				continue
			}

			asset, err := assetService.Process(line)
			if err != nil {
				fmt.Println(err.Error())
			} else {
				assetService.Calculate(asset, cache)
			}
		}

		assetProcessedResults := assetService.ProcessResults(cache)
		for _, assetProcessedResult := range assetProcessedResults {
			b, err := json.Marshal(assetProcessedResult)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(string(b))
		}

	}(assetService, cache)

	cmd := exec.Command("./Binaries stdoutinator_amd64_linux.bin")
	cmd.Stdout = writer
	_ = cmd.Start()
	go func() {
		_ = cmd.Wait()
		cmdDone()
		writer.Close()
	}()
	<-cmdCtx.Done()

	<-scannerStopped
}
