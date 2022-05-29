package asset

type AssetProcessedResult struct {
	Market                     int     `json:"market"`
	TotalVolume                float64 `json:"total_volume"`
	MeanPrice                  float64 `json:"mean_price"`
	MeanVolume                 float64 `json:"mean_volume"`
	VolumeWeightedAveragePrice float64 `json:"volume_weighted_average_price"`
	PercentageBuy              float64 `json:"percentage_buy"`
}

type AssetProcessTotal struct {
	TotalVolume    float64
	TotalPrice     float64
	TotalBuy       int
	Count          int
	PricePerVolume float64
}
