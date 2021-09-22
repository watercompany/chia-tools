package scraper

type ScraperCfg struct {
	SrcDir             string
	DestDir            string
	Save               bool
	Print              bool
	Proofs             bool
	TotalPlots         bool
	TotalEligiblePlots bool
}
