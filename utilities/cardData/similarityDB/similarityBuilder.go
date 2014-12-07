package similarityBuilder
//exposes card similarity. Relies on AllCards-x.json

import(

	"os"
	"log"
	"fmt"

	//we call the GC regularly so memory doesn't get horribly abused
	"runtime"

	//statistics for time per round is powerful
	"time"

	"strings"

	"io/ioutil"
	"encoding/json"

	"sort"

	"io"

)

func getLogger(fName, name string) (aLogger *log.Logger) {
	file, err:= os.OpenFile(fName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err!=nil {
		fmt.Println("Starting logger failed, cannot write to logger to say logger failed. Oh god.")
		fmt.Println(err)
		os.Exit(0)
	}

	multi:= io.MultiWriter(file, os.Stdout)

	aLogger = log.New(multi, name, log.Ldate|log.Ltime|log.Lshortfile)

	return
}

//ngram sizes per round and the weights for those rounds
var rounds = [...]int{3,4,5,6,7,8,9,10,11, 12, 13, 14}
var weighting = [...]float64{1.0, 1.15, 1.3,1.6,2.0, 2.5, 3.5, 5.0, 6.5, 8.5, 10.0, 12.0}

//characters that we remove in card text to normalize data
var badCharacters = [...]string{".", ",", "\"", "{", "}", "[", "]"}

//how many cards similar to this card we provide
const relatedCardsPerCard int = 8

const cacheFile string = "similarityData.cache.json"

type QueryableSimilarityData struct{

	//how many choices computed we made available
	//per card
	ChoicesPerCard int

	OverallConfidence float64

	//the meat to the time
	Results map[string]*cardSimilarityResult

}

type cardSimilarityResult struct{

	//the card's actual name
	Name string

	//the cards most similar to this card
	Others []string

	//the confidence that the given card is legitimate
	//on a per choice basis. an index i maps to the index
	//i of others to be representative of the results
	Confidences []float64

	//because some cards, like planeswalkers, have a lot of text,
	//the absolute confidence isn't the most helpful thing.
	//
	//adjusted confidence tests against the most similar item
	//in terms of coverage.
	AdjustedConfidences []float64

}

func GetQueryableSimilarityData() QueryableSimilarityData {
	someData:= QueryableSimilarityData{}

	aLogger:= getLogger("similarityBuilder.log", "similarityBuilder")

	//see if a cache exists and we can read from it
	cacheData, err := ioutil.ReadFile(cacheFile)	
	if err!=nil {
		aLogger.Println("No cache available, creating")
	
		//manually compute the similarity data
		someData.populate(aLogger)

		//send it to cache
		cacheData, err = json.Marshal(someData)
		if err == nil{
			ioutil.WriteFile(cacheFile, cacheData, 0666)
		}

	}else{
		//restore from cache
		err = json.Unmarshal(cacheData, &someData)
		if err!=nil {
			aLogger.Fatalf("Failed to read from similarity cache")
		}

	}

	return someData
}

//returns the cardSimilarityResult for a card if possible, otherwise returns
//an error
func (someData *QueryableSimilarityData) Query(name string) (cardSimilarityResult,
	error) {
	
	usableName:= normalizeCardName(name)

	result, ok:= someData.Results[usableName]
	if !ok {
		return cardSimilarityResult{}, fmt.Errorf("Card Not Found")
	}

	return *result, nil

}

//populates a QueryableSimilarityData with similarity data. The majority of the
//work is cached between runs.
func (someData *QueryableSimilarityData) populate(aLogger *log.Logger) {

	//grab our cards with slight pre-processing
	cardMap:= acquireCardMap(aLogger)

	roundsStartTime:= time.Now()

	//run our ngram rounds on the cards
	ngramRoundResults:= ngramRounds(cardMap, aLogger)

	roundsEndTime:= time.Now()
	aLogger.Println("Running rounds took ", roundsEndTime.Sub(roundsStartTime))
	aLogger.Println("Building confidence normalized card map")


	results, overallConfidence:= confidenceNormalize(ngramRoundResults)

	someData.Results = results
	someData.OverallConfidence = overallConfidence
	someData.ChoicesPerCard = relatedCardsPerCard

}

func confidenceNormalize(cardToCardToCardCount map[string]map[string]float64) (map[string]*cardSimilarityResult,
 float64) {

	normalized:= make(map[string]*cardSimilarityResult)

	var runningConfidence float64

	for aCardName, relatedCards:= range cardToCardToCardCount{

		closestCards:= getSortedCount(relatedCards, relatedCardsPerCard)

		//we need a space at the front to record confidence
		namesOnly := make([]string, relatedCardsPerCard )
		floatsOnly := make([]float64, relatedCardsPerCard)
		adjustedFloatsOnly := make([]float64, relatedCardsPerCard)

		//the confidence of a given card's relation is by how close
		//the closest card is to the actual card. the actual card
		//occupies the zeroth spot
		var confidence float64

		for i, aCardSimilarityMeasure := range closestCards{

			if len(closestCards) < 3 {
				confidence = 0.0
			}else{
				confidence = closestCards[i].Count / closestCards[0].Count
				runningConfidence+= confidence

			}

			//offset by 1 so we can have a confidence rating in the front
			namesOnly[i] = aCardSimilarityMeasure.Name
			floatsOnly[i] = confidence

			if i>=1 {
					//the first element is the actual card,
					//we don't want to adjust relative to that for obvious
					//reasons

				if floatsOnly[1] == 0.0 {
					adjustedFloatsOnly[i] = 0.0
				}else{
					adjustedFloatsOnly[i] = confidence / floatsOnly[1]
				}

			}else{
				adjustedFloatsOnly[i] = 1.0
			}

		}

		normalized[aCardName] = &cardSimilarityResult{
					Name: aCardName,
					Others: namesOnly,
					Confidences: floatsOnly,
					AdjustedConfidences: adjustedFloatsOnly,
			}
	}

	totalConfidence := runningConfidence / float64(len(cardToCardToCardCount))


	return normalized, totalConfidence

}

type cardSimilarityMeasures []cardSimilarityMeasure

type cardSimilarityMeasure struct{
	Name string
	Count float64
}

func cardSimilarityMeasuresFromMap(mappedCount map[string]float64) cardSimilarityMeasures {
	
	someMeasures := make(cardSimilarityMeasures, len(mappedCount))


	i := 0
	for aName, aCount := range mappedCount{
		someMeasures[i] = cardSimilarityMeasure{aName, aCount}
	
		i++
	}

	return someMeasures

}

func (someData cardSimilarityMeasures) Len() int {
	return len(someData)
}

func (someData cardSimilarityMeasures) Less(i, j int) bool {
	return someData[j].Count < someData[i].Count
}

func (someData cardSimilarityMeasures) Swap(i, j int) {
	someData[i], someData[j] = someData[j], someData[i]
}

func getSortedCount(relatedCards map[string]float64,
	topToReport int) ([]cardSimilarityMeasure) {

	someMeasures := cardSimilarityMeasuresFromMap(relatedCards)

	sort.Sort(someMeasures)

	largestReportable := topToReport
	if largestReportable > len(someMeasures) {
		largestReportable = len(someMeasures)
	}

	topSection := make([]cardSimilarityMeasure, largestReportable)

	for i := 0; i < largestReportable; i++ {			
		
		topSection[i] = someMeasures[i]

	}

	return topSection
	
}

//returns an array of maps mapping from card name to cards similar to that card
//which then maps to the float representing the confidence that these cards
//are similar
func ngramRounds(cardMap map[string]card, 
	aLogger *log.Logger) map[string]map[string]float64 {
	
	roundResults := make([]map[string]map[string]float64, len(rounds))

	for i, aRoundSize := range rounds{

		startTime := time.Now()

		weight := weighting[i]

		aLogger.Println("Starting round ", i+1, " with weight ", weight,
			" n = ", aRoundSize)

		roundResults[i] = ngramMapRound(cardMap, aRoundSize, weight, aLogger)

		runtime.GC()

		endTime := time.Now()
		aLogger.Println("Time for this round is ", endTime.Sub(startTime))

	}

	//with the similarity per round tallied, we need to determine
	//similarity holistically, between all rounds

	//using the first set of results as a basis 
	baseResultSet := roundResults[0]

	for i, aRoundResult := range roundResults{

		if i == 0 {
			//we already used the first round as the base
			continue
		}

		for referenceCardName, relatedCards := range aRoundResult{

			for aSimilarCardName, similarityCount := range relatedCards{

				baseResultSet[referenceCardName][aSimilarCardName]+= similarityCount

			}

		}

	}

	return baseResultSet

}

func ngramMapRound(cardMap map[string]card, n int,
	weight float64, aLogger *log.Logger) map[string]map[string]float64  {
	
	startTime := time.Now()

	cardToNgrams, nGramMap := cardMapToNgramMap(cardMap, n)

	postnGramTime := time.Now()

	aLogger.Println("nGrams found, time taken is ", postnGramTime.Sub(startTime))

	cardToCardToCardCount:= ngamMapToCardCounts(cardToNgrams, nGramMap, weight)

	aLogger.Println("Counting time is ", time.Now().Sub(postnGramTime))

	return cardToCardToCardCount

}

func cardMapToNgramMap(cardMap map[string]card, n int) (map[string][]string, map[string][]string) {

	//we have the card name to the trigrams
	cardToNgrams:= make(map[string][]string)

	//then the trigrams mapped to the matching card names
	nGramMap := make(map[string][]string)
	
	for cardName, aCard:= range cardMap{

		splitText:= strings.Split(aCard.Text, "\n")

		for _, someText := range splitText{

			goodText:= strings.Split(someText, " ")

			k:= 0
			limit := len(goodText) - (n-1)
			for k < limit{

				anNgram := strings.Join(goodText[k:(k+n)], " ")

				nGramMap[anNgram] = append(nGramMap[anNgram], cardName)

				cardToNgrams[cardName] = append(cardToNgrams[cardName], anNgram)

				k++
			}

		}

	}

	return cardToNgrams, nGramMap

}

func ngamMapToCardCounts(cardToNgrams, nGramMap map[string][]string,
	weight float64) map[string]map[string]float64 {
	//now that we have a map of ngrams to cards
	//and a map of card to ngrams, we merge the two by
	//creating a map of card names to a map of card names to
	//ints with which we calculate shared ngrams between each card name

	cardToCardToCardCount := make(map[string]map[string]float64)

	for cardName, nGrams := range cardToNgrams{

		cardToCardToCardCount[cardName] = make(map[string]float64)


		for _, anNgram := range nGrams{
			//this is the individual ngrams level

			sharedCards := nGramMap[anNgram]

			//we remove the most common sets of ngrams by having
			//an inverse relation between quantity of shared cards and
			//weight of that specific ngrams.
			
				//optimize this a little
			var value float64 = (weight / float64(len(sharedCards)))
 
			for _, aSharedCardName := range sharedCards{
				cardToCardToCardCount[cardName][aSharedCardName]+= value
			}


		}

	}


	return cardToCardToCardCount
}

type cardMap map[string]card

//the components of the card we use for similarity determination
type card struct{
	Name string
	Text string
	Power, Toughness string
	Printings, Types []string
}

func acquireCardMap(aLogger *log.Logger) (map[string]card) {
	
	aLogger.Println("Acquiring raw card data")

	//grab the card data hosted on disk
	cardData, err:= ioutil.ReadFile("AllCards-x.json")
	if err!=nil {
		aLogger.Fatalf("Failed to read AllCards-x.json")
	}

	//unmarshal it into a map of string to card with relevant data
	var aCardList cardMap
	err = json.Unmarshal(cardData, &aCardList)
	if err!=nil {
		aLogger.Fatalf("Failed to unmarshal card list")
	}

	aLogger.Println("Raw card map unmarshalled")

	cleanedCardMap := make(map[string]card)
	for _, aCard := range aCardList{

		//ensure this card isn't of a type we ignore

		cleanedCardMap[aCard.Name] = *processCard(&aCard)

	}

	aLogger.Println("Raw card map acquired")

	return cleanedCardMap

}

//processes a card for ngram usage
func processCard(aCard *card) *card {
	
	//clean the card's text
	aCard.Text = cleanCardText(aCard.Text, aCard.Name)

	//remove the name
	aCard.Name = normalizeCardName(aCard.Name)

	//add some contextually helpful data

	//we append the regular and sub types of the card to the text we'll be analyzing
	aCard.Text += "\n" + "T: " + strings.Join(aCard.Types, " ")

	//and we'll include power and toughness	//
	//we mold it very carefully to keep these in the trigram/bigram level
	aCard.Text+= "\n" +  aCard.Power + " "  + aCard.Toughness

	return aCard

}

func cleanCardText(someText, name string) string {
		//we need to clean the text of the string for hard brackets
	cleanedText:= someText

	//first, switch the name out for a token
	cleanedText = strings.Replace(cleanedText, name, "TheName", -1)

	openerIndex := strings.Index(cleanedText, "(")
	closerIndex := strings.Index(cleanedText, ")")

	for openerIndex!=-1 &&
		closerIndex!=-1{

		cleanedText = strings.Replace(cleanedText, cleanedText[openerIndex:closerIndex+1], "", 1)

		openerIndex = strings.Index(cleanedText, "(")
		closerIndex = strings.Index(cleanedText, ")")

	}

	cleanedText = strings.ToLower(cleanedText)

	for _, aBadChar := range badCharacters{
		cleanedText = strings.Replace(cleanedText, aBadChar, "", -1)
	}

	cleanedText = strings.TrimSpace(cleanedText)

	return cleanedText
}

func normalizeCardName(name string) string {
	
	if strings.Index(name, "Æ")!=-1 {
		name = strings.Replace(name, "Æ", "AE", -1)
	}

	return name

}