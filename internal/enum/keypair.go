package enum

type KeypairAlgorithm uint8

const (
	KeypairAlgoRS256 KeypairAlgorithm = 1 + iota
	KeypairAlgoES256
)

type KeypairStatus uint8

const (
	KeypairStatusActive KeypairStatus = 1 + iota
	KeypairStatusGrace
	KeypairStatusRetired
)
