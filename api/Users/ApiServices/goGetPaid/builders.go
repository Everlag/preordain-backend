package getPaid

import(

	"github.com/stripe/stripe-go"

	"encoding/json"
	"io/ioutil"

)

func GetMerchant(key string) *Merch {

	// Set the global stripe key
	stripe.Key = key

	return &Merch{}
	
}

func GetMerchantFromFile(loc string) (*Merch, error) {
	metaRaw, err:= ioutil.ReadFile(loc)
	if err!=nil {
		return nil, err
	}

	var meta MerchMeta
	err = json.Unmarshal(metaRaw, &meta)
	if err!=nil {
		return nil, err	
	}

	merch:= GetMerchant(meta.PrivateKey)

	return merch, nil
}

type MerchMeta struct{

	PrivateKey string

}