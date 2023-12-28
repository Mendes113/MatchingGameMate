
## Game Recommendation System

### Overview
This Go program is designed to retrieve and analyze game data using the RAWG Video Games Database API, store the information in a MongoDB database, and provide game recommendations based on user preferences.

### Features
1. **Data Retrieval from RAWG API:**
   - The program fetches game data from the RAWG API using an API key for authentication.
   - The retrieved data includes information about games, such as name, genres, and rating.

2. **MongoDB Integration:**
   - Game data is stored in a MongoDB database named "gamestore" within the "games" collection.

3. **User Preferences:**
   - User preferences are represented by the `UserChoices` struct, including a username and a list of preferred game genres.

4. **User Comparison:**
   - The program can compare the game preferences of two users (`UserChoices`) and identify games that match their combined preferences.

5. **Top Rated Games:**
   - The program can identify the top 5 games from a given list based on their ratings.

6. **Web Scraping for Similar Games:**
   - The program utilizes headless browser automation (Chromedp) to scrape a website (`https://gameslikefinder.com/`) for games similar to the top-rated games.

### Setup

1. **Dependencies:**
   - Ensure that you have Go installed on your machine.
   - Install the required Go packages using:
     ```bash
     go get -u github.com/chromedp/cdproto/cdp
     go get -u github.com/chromedp/chromedp
     go get -u go.mongodb.org/mongo-driver/bson
     go get -u go.mongodb.org/mongo-driver/mongo
     ```

2. **MongoDB:**
   - Install and run MongoDB locally.
   - Update the MongoDB connection URI in the `mongoConnection` function if needed.

3. **API Key:**
   - Obtain an API key from the RAWG Video Games Database API (https://rawg.io/apidocs).

4. **Code Configuration:**
   - Replace the placeholder API key (`const apiKey = "YOUR_API_KEY"`) with your actual API key.

### Usage

1. **Data Retrieval and Storage:**
   - Uncomment the relevant code in the `main` function to retrieve game data from the RAWG API and store it in the MongoDB database.

2. **User Preferences and Recommendations:**
   - Define user preferences using the `UserChoices` struct and call the `equalGamesIn2Users` and `top5GamesRated` functions to get game recommendations.

3. **Web Scraping for Similar Games:**
   - Uncomment the relevant code in the `main` function to perform web scraping for similar games.

### Notes
- The program includes functions to filter games by genre list, retrieve games from MongoDB, and establish a MongoDB connection.
- Adjustments may be needed based on changes in the RAWG API or the structure of the target website for web scraping.

### Disclaimer
This code is provided as a sample and may require further customization based on specific use cases and requirements. Use it responsibly and in accordance with the terms of service of the APIs and websites involved.
