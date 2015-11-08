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

// Updates a given customer to the provided plan.
//
// Requires the customer's id, sub id, and the plan id
func (merch *Merch) UpdateSubCustomer(customerID, subID, plan string) error {
	
	subParams:= &stripe.SubParams{
		Customer: customerID,
		Plan: plan,
	}

	_, err := sub.Update(subID, subParams)

	return err

}

// Removes a given customer's subscription
//
// NOTE: updating a subscription should use the dedicated update
// method as that leverages stripe's prorating
//
// Requires both the customer's id and their accompanying
// subscription id
func (merch *Merch) UnSubCustomer(subID, customerID string) error {

	subParams:= &stripe.SubParams{
		Customer: customerID,
	}

	return sub.Cancel(subID, subParams)
}

// Adds a new customer with a given email and payment token.
//
// token must be a stripe provided token.
//
// Returns a valid customer id if successful
func (merch *Merch) AddCustomer(token, email, coupon string) (string, error) {

	customerParams := &stripe.CustomerParams{
	  Email: email,
	  Coupon: coupon,
	  Source: &stripe.SourceParams{
	  	Token: token,
	  	},
	}

	// Send the new customer off
	c, err := customer.New(customerParams)
	if err!=nil {
		return "", err
	}

	return c.ID, nil

}

// Updates a customer to a new payment token.
//
// Requires a new payment token to take the place of the one
// used in the previous subscription.
func (merch *Merch) UpdateCustomer(customerID, token string) error {
	customerParams:= &stripe.CustomerParams{
		Source: &stripe.SourceParams{
	  		Token: token,
	  	},
	}

	_, err := customer.Update(customerID, customerParams)

	return err

}