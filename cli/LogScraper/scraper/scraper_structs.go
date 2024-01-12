package scraper

type ScraperCfg struct {
	SrcDir       string
	DestDir      string
	Save         bool
	Print        bool
	SendTelegram bool
	BotToken     string
	ChatID       string

	StartDate string

	Proofs             bool
	TotalProofsFound   bool
	TotalPlots         bool
	TotalEligiblePlots bool
	MaxProofTime       bool
	MedianProofTime    bool
	MeanProofTime      bool
	PercentProofTime   int
	GapsProofChecks    int

	TotalProofsFoundInt int
}

type FarmDateMap struct {
	FarmIndex int
	Date      string
}
