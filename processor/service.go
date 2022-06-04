package processor

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/leonardyeoxl/messari-market-data-coding-challenge/asset"
)

type IProcessorService interface {
	Initialize(path string, writer *io.PipeWriter) (*exec.Cmd, error)
	Start(cmd *exec.Cmd, cmdDone context.CancelFunc, writer *io.PipeWriter)
	Read(assetService asset.IAssetService, reader *io.PipeReader, scannerStopped chan struct{}, assetProcessedChannel chan []asset.AssetProcessedResult)
	WriteOutput(path string, assetProcessedResults []asset.AssetProcessedResult) error
}

type ProcessorService struct{}

func NewProcessorService() IProcessorService {
	return &ProcessorService{}
}

func (ps ProcessorService) Initialize(path string, writer *io.PipeWriter) (*exec.Cmd, error) {
	cmd := exec.Command(path)
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
	assetProcessedChannel chan []asset.AssetProcessedResult,
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
			log.Fatal(err.Error())
		} else {
			assetService.Calculate(asset, cache)
		}
	}

	assetProcessedResults := assetService.ProcessResults(cache)
	assetProcessedChannel <- assetProcessedResults
	// ps.WriteOutput("./data/processed/results.json", assetProcessedResults)
}

func (ps ProcessorService) WriteOutput(
	path string,
	assetProcessedResults []asset.AssetProcessedResult,
) error {
	file, _ := json.MarshalIndent(assetProcessedResults, "", " ")
	err := ioutil.WriteFile(path, file, 0644)
	if err != nil {
		return err
	}
	return nil
}
