package magiccardmarket

type productResponse struct {
	Product struct {
		// What we are looking for
		PriceGuide struct {
			AVG     float64 `json:"AVG"`
			LOW     float64 `json:"LOW"`
			LOWEX   float64 `json:"LOWEX"`
			LOWFOIL float64 `json:"LOWFOIL"`
			SELL    float64 `json:"SELL"`
			TREND   float64 `json:"TREND"`
		} `json:"priceGuide"`
		// Where the listing on MKM is located.
		Website string `json:"website"`

		Name struct {
			English struct {
				IdLanguage   int    `json:"idLanguage"`
				LanguageName string `json:"languageName"`
				ProductName  string `json:"productName"`
			} `json:"1"`
		} `json:"name"`

	} `json:"product"`
}

// For an expansion, we just want to get the ids of the product inside it,
// querying the product itself will reveal to us the name of the product.
type expansionResponse struct {
	Card []struct {
		IdProduct int `json:"idProduct"`
	} `json:"card"`
}

type setListResponse struct{
	Expansion []setIdentity
}

type setIdentity struct{
	IdExpansion int
	Name string
	CleanedName string
}

// We set the name of the identity to be clean in the event that a cleaned
// name exists for this set
func (anIdentity *setIdentity) setProperName() {
	cleanName, ok:= dirtyToCleanSetNames[anIdentity.Name]
	if ok {
		anIdentity.CleanedName = cleanName
	}else{
		anIdentity.CleanedName = anIdentity.Name
	}
}
