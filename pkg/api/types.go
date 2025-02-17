package celestials

type Changes struct {
	Head    int64    `json:"head"`
	Changes []Change `json:"changes"`
}

type Change struct {
	CelestialID string `json:"celestial_id"`
	Address     string `json:"address"`
	ImageURL    string `json:"image_url,omitempty"`
	ChangeID    int64  `json:"change_id"`
	Status      string `json:"status"`
}
