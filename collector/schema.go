package collector

type ClustsafeSchema struct {
	Modules struct {
		Clustsafe []struct {
			ID      string `xml:"id,attr"`
			Status  string `xml:"status"`
			Power   Power  `xml:"power"`
			Outlets struct {
				Count  string `xml:"count,attr"`
				Outlet []struct {
					ID        string `xml:"id,attr"`
					Status    string `xml:"status"`
					Fuse      string `xml:"fuse"`
					Autopower string `xml:"autopower"`
					Power     Power  `xml:"power"`
				} `xml:"outlet"`
			} `xml:"outlets"`
			Lines struct {
				Count string `xml:"count,attr"`
				Line  []struct {
					ID             string `xml:"id,attr"`
					Status         string `xml:"status"`
					Identification string `xml:"identification"`
					Power          Power  `xml:"power"`
				} `xml:"line"`
			} `xml:"lines"`
		} `xml:"clustsafe"`
	} `xml:"modules"`
	Sensors struct {
		Count  string `xml:"count,attr"`
		Sensor []struct {
			Type       string  `xml:"type,attr"`
			ID         string  `xml:"id,attr"`
			Status     string  `xml:"status"`
			Value      float64 `xml:"value"`
			Alert      string  `xml:"alert"`
			Identifier string  `xml:"identifier"`
		} `xml:"sensor"`
	} `xml:"sensors"`
}

type Power struct {
	Status        string  `xml:"status,attr"`
	Voltage       float64 `xml:"voltage"`
	Current       float64 `xml:"current"`
	Frequency     float64 `xml:"frequency"`
	RealPower     float64 `xml:"realPower"`
	ApparentPower float64 `xml:"apparentPower"`
	PowerFactor   float64 `xml:"powerFactor"`
	PhaseShift    float64 `xml:"phaseShift"`
	Samples       float64 `xml:"samples"`
}
