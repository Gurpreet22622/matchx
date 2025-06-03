package models

type RegisterProperty struct {
	ID              string      `json:"id"`
	UserID          string      `json:"user_id"`
	PropertyType    string      `json:"property_type"`
	Location        Coordinates `json:"location"`
	Locality        string      `json:"locality"`
	LeaseType       string      `json:"lease_type"`
	FurnishedStatus string      `json:"furnished_status"`
	PropertyArea    float32     `json:"property_area"`
	Internet        bool        `json:"internet"`
	AC              bool        `json:"ac"`
	RO              bool        `json:"ro"`
	Kitchen         bool        `json:"kitchen"`
	Geezer          bool        `json:"geezer"`
	Rent            float32     `json:"rent"`
}
