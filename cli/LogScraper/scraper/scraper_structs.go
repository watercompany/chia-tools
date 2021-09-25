package scraper

type ScraperCfg struct {
	SrcDir  string
	DestDir string
	Save    bool
	Print   bool

	Proofs             bool
	TotalPlots         bool
	TotalEligiblePlots bool
	MaxProofTime       bool
	MedianProofTime    bool
	MeanProofTime      bool
}

type FarmDateMap struct {
	FarmIndex int
	Date      string
}
