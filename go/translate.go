package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"gopkg.in/resty.v1"

	"github.com/algolia/algoliasearch-client-go/algoliasearch"
	"github.com/joho/godotenv"
)

func translate(textToTranslate, algoliaObjectID, targetLanguage string) string {

	// Create a struct for the JSON we need to send to IBM
	type TranslationObject struct {
		Text   string `json:"text,omitempty"`
		Source string `json:"source,omitempty"`
		Target string `json:"target,omitempty"`
	}

	// Create an JSON object using the above struct
	body := &TranslationObject{
		Text:   textToTranslate,
		Source: "en",
		Target: targetLanguage,
	}

	// Send the JSON object to IBM for translation
	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetBasicAuth(os.Getenv("IBM_USERNAME"), os.Getenv("IBM_PASSWORD")).
		SetBody(body).
		Post("https://gateway.watsonplatform.net/language-translator/api/v2/translate")

	// If there's no errors, map the response to JSON
	if err == nil {
		var result map[string]interface{}
		json.Unmarshal(resp.Body(), &result)

		translationsArray := result["translations"].([]interface{})
		translations := translationsArray[0].(map[string]interface{})
		translation := translations["translation"].(string)

		// Prepare a new object for updating
		// Dynamically set the description_:lang
		objectToUpdate := algoliasearch.Object{
			"objectID": algoliaObjectID, "description_" + targetLanguage: translation,
		}

		// Init Algolia
		client := algoliasearch.NewClient(os.Getenv("AG_APPLICATION_ID"), os.Getenv("AG_API_WRITE_KEY"))
		index := client.InitIndex(os.Getenv("AG_INDEX_NAME"))
		_, err := index.PartialUpdateObject(objectToUpdate)

		if err == nil {
			fmt.Println("Translation added to " + algoliaObjectID)
		} else {
			fmt.Println("There was an error adding translated field to " + algoliaObjectID + ": " + string(err.Error()))
		}

	}

	return ""
}

func main() {
	// Load the .env
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env")
	}

	// Grab the objectID
	algoliaObjectID := os.Args[1]

	// Grab the target language (en, fr, de, es)
	targetLanguage := os.Args[2]

	// Init Algolia
	client := algoliasearch.NewClient(os.Getenv("AG_APPLICATION_ID"), os.Getenv("AG_API_WRITE_KEY"))
	index := client.InitIndex(os.Getenv("AG_INDEX_NAME"))

	// Get the object we need for translation from Algolia
	object, err := index.GetObject(algoliaObjectID, []string{"description_en"})

	// Assign the returned field from Algolia to a variable
	text := object["description_en"].(string)

	// Let us know what's happening
	fmt.Println("Retrieved data for " + algoliaObjectID + ". Waiting for translation...")

	// Pass the text and target language over for translation
	translate(text, algoliaObjectID, targetLanguage)

}
