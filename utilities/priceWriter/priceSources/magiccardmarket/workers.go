package magiccardmarket

import(

	"fmt"

	"strings"

	"net/http"

	"sync"
)

// Populates the provided price map concurrently using WorkerCount number
// of workers.
//
// Returns an error if it encountered a error. Otherwise, returns
// nil. Due to the concurrent nature, if multiple errors occur across workers
// then there is absolutely no guarantee regarding odering of errors. 
func runWorkers(setList []string,
	priceMap map[string]map[string]int64,
	cleanToMKM map[string]string,
	consumerKey, consumerSecret string) error {

	// Create a buffered channel to feed the price workers set names.
	setChan:= make(chan string, len(setList))
	// and to receive operating errors
	errorChan:= make(chan error, WorkerCount + 1)

	// Create a way to wait on our workers
	completion:=  &sync.WaitGroup{}
	// Prvent concurrent access to price map
	priceMapLock:= &sync.Mutex{}

	// Add the sets
	for _, aSet:= range setList{
		setChan <- aSet
	}


	for i := 0; i < WorkerCount; i++ {

		// Populate the completion record
		completion.Add(1)

		go magiccardmarketPriceWorker(setChan, errorChan,
			completion,
			priceMap, priceMapLock,
			cleanToMKM,
			consumerKey, consumerSecret)

	}

	// Wait for each worker to perform.
	completion.Wait()

	// We try to rip the topmost error off the top. If an error exists,
	// we return it.
	select{

	case err:= <- errorChan:
		return err
	default:
		break

	}

	return nil
	
}

// A worker for acquiring MKM prices.
//
// Provided necessary credentials as well as a set feeder channel and
// a waitgroup, this will populate the provided priceMap, locking
// when necessary to atomize setting.
//
// Provide it a pointer to a place where it can store any encountered error.
// In the event an error is encountered, it will assign the error and exit.
//
// Creates its personal http client.
func magiccardmarketPriceWorker(setChan chan string, errorChan chan error,
	completion *sync.WaitGroup,
	priceMap map[string]map[string]int64, priceMapLock *sync.Mutex,
	cleanToMKM map[string]string,
	consumerKey, consumerSecret string) {

	// Always make sure we make it known that we have completed
	defer completion.Done()

	aClient:= &http.Client{}
	for{

		// When the channel is out of sets for us we exit
		select{

		case fullSetName:= <- setChan:

			err:= workerSetAcquisition(fullSetName, aClient,
				priceMap, priceMapLock,
				cleanToMKM, consumerKey, consumerSecret)
			if err!=nil {
				errorChan <- err
				return				
			}

		default:
			// Signal that we are done and bow out.
			return
		}

	}

}

func workerSetAcquisition( fullSetName string,
	aClient *http.Client,
	priceMap map[string]map[string]int64, priceMapLock *sync.Mutex,
	cleanToMKM map[string]string,
	consumerKey, consumerSecret string) error {

	// Massage the set name
	if fullSetName == "" {
		return nil
	}

	// Figure out if we need the foil flag.
	var foil bool
	aSet := fullSetName
	if strings.Contains(fullSetName, " Foil") {
		foil = true
		aSet = strings.Replace(fullSetName, " Foil", "", -1)
	}else{
		foil = false
	}

	if isIgnoredSetName(aSet){
		return nil
	}


	setPrices, err:= getExpansionPrices(aSet, foil,
		cleanToMKM,
		consumerKey, consumerSecret,
		aClient)
	if err!=nil {
		return fmt.Errorf("Encountered error fetching ",
			aSet, err)
	}

	// Atomically add the price to our results
	priceMapLock.Lock()
	priceMap[fullSetName] = setPrices
	priceMapLock.Unlock()

	return nil
}