package config

const (
	MaxDatasetBytes    = 50 * 1024 * 1024 // 50MB
	MaxImageBytes      = 5 * 1024 * 1024  // 5MB
	MaxImages          = 10
	MaxUploadBytes     = 50 * 1024 * 1024 // 50MB
	MaxVerifyBytes     = 5 * 1024 * 1024  // 5MB
	MaxMultipartMemory = 100 << 20        // 100MB for evaluate (dataset + images)
)
