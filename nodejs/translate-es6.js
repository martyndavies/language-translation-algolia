#!/usr/bin/node

if (process.env.NODE_ENV != 'production') require('dotenv').config();

// Set up Algolia
const algolia = require('algoliasearch');
const client = algolia(
  process.env.AG_APPLICATION_ID,
  process.env.AG_API_WRITE_KEY
);
const index = client.initIndex(process.env.AG_INDEX_NAME);

// Set up Axios
const axios = require('axios');

// Take in the objectID from the index as an argument
// Example: $node translate 3dvf4fg
const algoliaObjectID = process.argv[2];

// Take in the target language (es, fr, de)
const targetLanguage = process.argv[3];

// Translate the text
const translate = inputText => {
  console.log(
    'Retrieved data for ' + algoliaObjectID + '. Waiting for translation...'
  );

  // Create an object of the input text, source lang and target lang
  let objectToTranslate = {
    text: inputText, // The words to be translated
    source: 'en', // The language they are in
    target: targetLanguage // The language you want them to be
  };

  // Request the translation from IBM Watson
  axios({
    method: 'post',
    url:
      'https://gateway.watsonplatform.net/language-translator/api/v2/translate',
    data: objectToTranslate,
    headers: { Accept: 'application/json' },
    auth: {
      username: process.env.IBM_USERNAME,
      password: process.env.IBM_PASSWORD
    }
  })
    .then(response => {
      // The response object is long, let's shorten it.
      const { translations } = response.data;

      // Is there a translation?
      if (translations.length > 0 && typeof translations !== 'undefined') {
        let translatedObject = {
          objectID: algoliaObjectID
        };
        // This is a new field we'll add to the existing object
        // Here we're dynamically setting the key so that different language
        // translations can be added to the same object
        translatedObject['description_' + targetLanguage] =
          translations[0].translation;

        // If yes, send it to be added to our Algolia object
        index
          .partialUpdateObject(translatedObject)
          .then(content =>
            console.log('Translation added to ' + algoliaObjectID)
          )
          .catch(err =>
            console.log('Error adding translation to ' + algoliaObjectID)
          );
      } else {
        // If no translation then let us know that
        console.log('Error translating ' + algoliaObjectID);
      }
    })
    .catch(err => console.log(err));
};

index
  .getObject(algoliaObjectID)
  .then(content => translate(content.description_en))
  .catch(err => {
    throw err;
  });
