package storage

// swagger:enum Status
/*
	ENUM(
		NOT_VERIFIED,
		VERIFIED,
		PRIMARY
	)
*/
//go:generate go-enum --marshal --sql --values --names
type Status string
