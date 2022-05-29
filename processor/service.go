package processor

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

type IProcessorService interface {
	Initialize(writer *io.PipeWriter) (*exec.Cmd, error)
	Start(cmd *exec.Cmd, cmdDone context.CancelFunc, writer *io.PipeWriter)
	Read(assetService asset.IAssetService, reader *io.PipeReader, scannerStopped chan struct{})
}

type ProcessorService struct{}

func NewProcessorService() IProcessorService {
	return &ProcessorService{}
}

func (ps ProcessorService) Initialize(writer *io.PipeWriter) (*exec.Cmd, error) {
	cmd := exec.Command("./data/raw/Binaries stdoutinator_amd64_linux.bin")
	cmd.Stdout = writer
	err := cmd.Start()
	if err != nil {
		return cmd, err
	}
	return cmd, nil
}

func (ps ProcessorService) Start(
	cmd *exec.Cmd,
	cmdDone context.CancelFunc,
	writer *io.PipeWriter,
) {
	_ = cmd.Wait()
	cmdDone()
	writer.Close()
}

func (ps ProcessorService) Read(
	assetService asset.IAssetService,
	reader *io.PipeReader,
	scannerStopped chan struct{},
) {
	defer close(scannerStopped)

	cache := map[int]asset.AssetProcessTotal{}

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
}
