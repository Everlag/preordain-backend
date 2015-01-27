package imageScraper
//scrapes card fulls and crops according to the mtgjson AllSets-x.json
//provided
//
//all references to cardNames are to their image name that mtgimage
//graciously offers

import(
	
	"os"
	"log"
	"fmt"

	"io/ioutil"
	"encoding/json"

	"io"
	"net/http"

)

//how many scrapers we can have running at once.
//
//networks blocks are our most likely issue
const maxGatherers int = 4

//the basic data so we know where to acquire, and how to save, our images
const imageBaseUrl string = "http://mtgimage.com/card/"
const cropExtension string = ".crop.jpg"
const fullExtension string = ".jpg"

//the basic data so we know where to acquire those beautiful svgs
//
//all symbols are found at symbol'MODIFIER'BaseUrl/'symbolNAME'.svg
const symbolExtension string = ".svg"
const symbolOtherBaseUrl string = "http://mtgimage.com/symbol/other/"
var symbolOtherList = [...]string{
	"c", "chaos", "chaosdice", "creature",
	"enchantment", "forwardslash", "instant", "land",
	"multiple", "planeswalk", "planeswalker", "power",
	"q", "sorcery", "t", "tap", "toughness", "untap",	
}
const symbolManaBaseUrl string = "http://mtgimage.com/symbol/mana/"
//programmatically scraped, newlines added manually for readability
var symbolManaList = [...]string{
	"0","1","10","100","1000000","11","12","13","14","15","16","17","18","19",
	"2","20","2b","2black","2blue","2g","2green","2r","2red","2u","2w","2white",
	"3","4","5","6","7","8","9","b","b2","bg","bh","bhalf","black","black2",
	"blackblue","blackgreen","blackh","blackhalf","blackp","blackphyrexian",
	"blackred","blacktwo","blackwhite","blue","blue2","blueblack","bluegreen",
	"blueh","bluehalf","bluep","bluephyrexian","bluered","bluetwo","bluewhite",
	"bp","bphyrexian","br","bu","bw","colorlesshalf","eight","eighteen","eleven",
	"fifteen","five","four","fourteen","g","g2","gb","gh","ghalf","gp",
	"gphyrexian","gr","green","green2","greenblack","greenblue","greenh",
	"greenhalf","greenp","greenphyrexian","greenred","greentwo","greenwhite",
	"gu","gw","h","half","halfb","halfblack","halfblue","halfcolorless","halfg",
	"halfgreen","halfr","halfred","halfu","halfw","halfwhite","hb","hg","hr",
	"hu","hundred","hw","infinity","million","nine","nineteen","one",
	"onehundred","onemillion","p","pb","pblack","pblue","pg","pgreen",
	"phyrexian","phyrexianblack","phyrexianblue","phyrexiangreen",
	"phyrexianred","phyrexianwhite","pr","pred","pu","pw","pwhite","r","r2",
	"rb","red","red2","redblack","redblue","redgreen","redh","redhalf",
	"redp","redphyrexian","redtwo","redwhite","rg","rh","rhalf","rp",
	"rphyrexian","ru","rw","s","seven","seventeen","six","sixteen","snow",
	"ten","thirteen","three","twelve","twenty","two","twoblack","twoblue",
	"twogreen","twored","twowhite","u","u2","ub","ug","uh","uhalf","up",
	"uphyrexian","ur","uw","w","w2","wb","wg","wh","whalf","white","white2",
	"whiteblack","whiteblue","whitegreen","whiteh","whitehalf","whitep",
	"whitephyrexian","whitered","whitetwo","wp","wphyrexian","wr","wu","x",
	"y","z","zero","∞",
}

const symbolSetBaseUrl string = "http://mtgimage.com/symbol/set/"

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

// exists returns whether the given file or directory exists or not
func exists(path string) (bool, error) {
    _, err := os.Stat(path)
    if err == nil { return true, nil }
    if os.IsNotExist(err) { return false, nil }
    return false, err
}

type cardMap map[string]card

//the only component of each card that we require is the image
type card struct{
	ImageName string
}

//scrapes the images for the cards and sticks them in the provided folders
//which are relative to our current working directory
func ScrapeImages(fullLoc, cropLoc, symbolLoc string) {
	
	aLogger:= getLogger("imageLog.txt", "imageLog")

	//grab the card data hosted on disk
	cardData, err:= ioutil.ReadFile("AllCards-x.json")
	if err!=nil {
		aLogger.Fatalf("Failed to read AllCards-x.json")
	}

	//unmarshal it into a map of string to card with image name
	var aCardList cardMap
	err = json.Unmarshal(cardData, &aCardList)
	if err!=nil {
		aLogger.Fatalf("Failed to unmarshal card list")
	}

	//divide up the work
	imagesPerGatherer:= len(aCardList) / maxGatherers
	gathererFeeders := make([][]string, maxGatherers)
	imagesAssignedThisGatherer:= 0
	currentGathererFeeder:= 0
	for _, aCard:= range aCardList{

		gathererFeeders[currentGathererFeeder] = append(
			gathererFeeders[currentGathererFeeder],
			aCard.ImageName )

		imagesAssignedThisGatherer++
		if imagesAssignedThisGatherer>imagesPerGatherer {
			//reset the count
			imagesAssignedThisGatherer = 0

			//increment the gatherer we feed
			currentGathererFeeder++
		}
	}

	//get the workers going
	completionChan:= make(chan bool)
	for i := 0; i < maxGatherers; i++ {
		go getCardList(gathererFeeders[i], fullLoc, cropLoc,
		 completionChan, aLogger)
	}

	//grab the symbols, the other workers should take a lot longer
	//than these svg symbols
	getSymbols(symbolLoc, aLogger)

	//now let the gatherers do their jobs!
	//
	//we attempt to pop a bool off the completion chan until we have
	//each gatherer accounted for
	for i := 0; i < maxGatherers; i++ {
		_ = <- completionChan
	}

}

//allows parallel usage of card image gathering
func getCardList(cardNames []string, fullLoc, cropLoc string,
	 completionChan chan bool, aLogger *log.Logger){
	for _, aCardName := range cardNames {
		getCardImage(aCardName, fullLoc, cropLoc, aLogger)
	}

	completionChan <- true
}

func getCardImage(cardName, fullLoc, cropLoc string, aLogger *log.Logger) {
	
	fullURL:= imageBaseUrl + cardName + fullExtension
	cropURL:= imageBaseUrl + cardName + cropExtension
	fullPath:= fullLoc + cardName + fullExtension
	cropPath:= cropLoc + cardName + fullExtension

	//we make a quick check to see if we have both the full and the crop
	//already are present
	fullExists, _:= exists(fullPath)
	cropExists, _:= exists(cropPath)
	if fullExists && cropExists {
		return
	}

	getDumbImage(fullURL, fullPath, aLogger)
	getDumbImage(cropURL, cropPath, aLogger)

}

//acquires all symbols declared statically at the top
func getSymbols(symbolLoc string, aLogger *log.Logger) {
	
	aLogger.Println("Acquiring Symbols")

	var url string
	var path string

	//acquire the mana symbols first. Many of these are synonymous
	for _, aSymbol:= range symbolManaList{
		url = symbolManaBaseUrl + aSymbol + symbolExtension
		path = symbolLoc + aSymbol + symbolExtension

		getDumbImage(url, path, aLogger)
	}

	//now the other symbols that are much, much less meaty
	for _, aSymbol:= range symbolOtherList{
		url = symbolOtherBaseUrl + aSymbol + symbolExtension
		path = symbolLoc + aSymbol + symbolExtension

		getDumbImage(url, path, aLogger)
	}

	err:= getSetSymbols(symbolLoc, aLogger)
	if err!=nil {
		aLogger.Fatalf("Failed to get set symbols, ", err)
	}

	aLogger.Println("All symbols acquired")

}


type setMap map[string] set
type set struct{
	Name string
}

func getSetSymbols(symbolLoc string, aLogger *log.Logger) error {

	setData, err:= ioutil.ReadFile("AllSets-x.json")
	if err!=nil {
		return fmt.Errorf("Failed to read AllSets-x.json")
	}

	//unmarshal it into a map of string to card with image name
	var aSetMap setMap
	err = json.Unmarshal(setData, &aSetMap)
	if err!=nil {
		return fmt.Errorf("Failed to unmarshal card map")
	}

	var url string
	var path string
	for setCode, aSet := range aSetMap {
		url = symbolSetBaseUrl + setCode + "/" + "c" + symbolExtension
		path = symbolLoc + aSet.Name + symbolExtension
		getDumbImage(url, path, aLogger)
	}

	return nil

}

// Acquires an image dumbly. If it fails then it is logged.
func getDumbImage(url, outPath string, aLogger *log.Logger) {
	
	resp, err := http.Get(url)
	if err != nil {
		aLogger.Println("Failed to acquire ", url)
		return
	}
	defer resp.Body.Close()
	out, err := os.Create(outPath)
	if err != nil {
		aLogger.Println("Failed to acquire ", url)
		return
	}
	defer out.Close()
	io.Copy(out, resp.Body)

}