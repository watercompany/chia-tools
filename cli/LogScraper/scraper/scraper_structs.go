package scraper

type ScraperCfg struct {
	SrcDir  string
	DestDir string
	Save    bool
	Print   bool

	StartDate string

	Proofs             bool
	TotalPlots         bool
	TotalEligiblePlots bool
	MaxProofTime       bool
	MedianProofTime    bool
	MeanProofTime      bool
	PercentProofTime   int
	GapsProofChecks    int
}

type FarmDateMap struct {
	FarmIndex int
	Date      string
}
