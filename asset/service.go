package asset

import "encoding/json"

type IAssetService interface {
	Process(line string) (Asset, error)
	Calculate(asset Asset, cache map[int]AssetProcessTotal)
	ProcessResults(cache map[int]AssetProcessTotal) []AssetProcessedResult
}

type AssetService struct{}

func NewAssetService() IAssetService {
	return &AssetService{}
}

func (as AssetService) Process(line string) (Asset, error) {
	asset := Asset{}
	err := json.Unmarshal([]byte(line), &asset)
	if err != nil {
		return asset, err
	}
	return asset, nil
}

func (as AssetService) Calculate(asset Asset, cache map[int]AssetProcessTotal) {
	if assetCalculate, ok := cache[asset.Market]; ok {
		assetCalculate.Count += 1
		if asset.IsBuy {
			assetCalculate.TotalBuy += 1
		}
		assetCalculate.TotalPrice += asset.Price
		assetCalculate.TotalVolume += asset.Volume
		assetCalculate.PricePerVolume += asset.Price * asset.Volume
	} else {
		totalBuy := 0
		if asset.IsBuy {
			totalBuy = 1
		}

		cache[asset.Market] = AssetProcessTotal{
			TotalVolume:    asset.Volume,
			TotalPrice:     asset.Price,
			TotalBuy:       totalBuy,
			Count:          1,
			PricePerVolume: asset.Price * asset.Volume,
		}
	}
}

func (as AssetService) ProcessResults(cache map[int]AssetProcessTotal) []AssetProcessedResult {
	assetProcessedResults := make([]AssetProcessedResult, 0)
	for market, assetProcessed := range cache {
		assetProcessedResult := AssetProcessedResult{
			Market:                     market,
			TotalVolume:                assetProcessed.TotalVolume,
			MeanPrice:                  assetProcessed.TotalPrice / float64(assetProcessed.Count),
			MeanVolume:                 assetProcessed.TotalVolume / float64(assetProcessed.Count),
			VolumeWeightedAveragePrice: assetProcessed.PricePerVolume / assetProcessed.TotalVolume,
			PercentageBuy:              float64(assetProcessed.TotalBuy) / float64(assetProcessed.Count),
		}
		assetProcessedResults = append(assetProcessedResults, assetProcessedResult)
	}
	return assetProcessedResults
}
