package getPaid

import(
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/sub"
	"github.com/stripe/stripe-go/customer"
)

// A merchant that allows us the ability to charge
// users.
//
// Mostly just a dummy that allows us to retain state as an object.
type Merch struct{}

func GetMerchant(key string) *Merch {

	// Set the global stripe key
	stripe.Key = key

	return &Merch{}
	
}

// Subscribes a given customer to a plan.
//
// Customer must be an customer id provided by stripe.
// Plan must be a plan id chosen at plan creation time.
//
// Returns a valid subscription id is successful
func (merch *Merch) SubCustomer(customer, plan string) (string, error) {

	subParams:= &stripe.SubParams{
		Customer: customer,
		Plan: plan,
	}

	s, err := sub.New(subParams)
	if err!=nil {
		return "", err
	}

	return s.ID, nil

}

// Removes a given customer from their plan
//
// Requires both the customer's id and their accompanying
// subscription id
func UnSubCustomer(subID, customerID string) error {
	return sub.Cancel(
			subID,
			&stripe.SubParams{Customer: customerID},
			)
}

// Adds a new customer with a given email and payment token.
//
// token must be a stripe provided token.
//
// Returns a valid customer id if successful
func (merch *Merch) AddCustomer(token, email string) (string, error) {

	customerParams := &stripe.CustomerParams{
	  Email: email,
	  Token: token,
	}

	// Send the new customer off
	c, err := customer.New(customerParams)
	if err!=nil {
		return "", err
	}

	return c.ID, nil

}