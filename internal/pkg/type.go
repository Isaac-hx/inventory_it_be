package pkg

type PaginationMeta struct {
	Page      int
	Limit     int
	TotalData int
	TotalPage int
}

type OverviewData struct {
	TotalAsset          int `json:"total_asset,omitempty"`
	TotalAssetAssigned  int `json:"total_asset_assigned,omitempty"`
	TotalAssetAvailable int `json:"total_asset_available,omitempty"`
	TotalAssetRetired   int `json:"total_asset_retired,omitempty"`
}
