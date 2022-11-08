package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/Krognol/go-wolfram"
	"github.com/joho/godotenv"
	"github.com/shomali11/slacker"
	"github.com/tidwall/gjson"

	witai "github.com/wit-ai/wit-go/v2"
)

func printCommandEvents(analyticsChannel <-chan *slacker.CommandEvent) {
	// This function is for debugging, it prints out each event's info whenever called

	for event := range analyticsChannel {
		fmt.Println("Command Events")
		fmt.Println(event.Timestamp)
		fmt.Println(event.Command)
		fmt.Println(event.Parameters)
		fmt.Println(event.Event)
		fmt.Println()

	}

}

func main() {

	// Load up the environment variables
	// The env contains the slack bot API, Wolfram Alpha API, and the Wit AI API  keys
	godotenv.Load(".env")

	// Create the Slack bot loaded with the Slack Bot Token, and the Slack App Token using the dotenv package
	// and os package to load the environment variables into slacker (from slacker package)
	bot := slacker.NewClient(os.Getenv("SLACK_BOT_TOKEN"), os.Getenv("SLACK_APP_TOKEN"))

	// create a client to access the WitAI API loaded up with the API Token from environment variables
	client := witai.NewClient(os.Getenv("WIT_AI_TOKEN"))

	// create a client to access the WolframAlpha API loaded up with the Wolfram App ID from environment variables
	wolframClient := &wolfram.Client{AppID: os.Getenv("WOLFRAM_APP_ID")}

	// Writing command function for the bot so it can querythe message
	// First parameter is the what to message the bot in the slack channel, second parameter is the command definition
	// Example @<Bot-Name> query <message>, message is the question for the bot
	// Example @SlackBot query What are the Colors of the Rainbow?
	bot.Command("query - <message>", &slacker.CommandDefinition{
		Description: "Ask any question to send it to Wolfram Alpha",
		Examples:    []string{"What are the Colors of the Rainbow?"},
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			// This is what the user sends to the bot
			query := request.Param("message")
			// Pass the User's query straight into WitAI to get an answer to the user's question
			msg, _ := client.Parse(&witai.MessageRequest{
				Query: query,
			})

			// Convert the WitAI API's response into a proper JSON format
			data, _ := json.MarshalIndent(msg, "", "    ")
			dataString := string(data[:])

			// Get the needed value parsed from the Object returned by WitAPI so we can use it to send to Wolfram Alpha API
			value := gjson.Get(dataString, "entities.wit$wolfram_search_query:wolfram_search_query.0.value")
			// Convert the value to String because wolfram alpha api client can only accept strings
			answer := value.String()

			// Pass in the answer string into wolfram alpha api which returns the answer to the question stored in the res variable
			// wolfram.Metric means to count the time it takes and 1000 is the time it takes to timeout
			res, err := wolframClient.GetSpokentAnswerQuery(answer, wolfram.Metric, 1000)
			// If an error has occured, log the error (with Fatal flag)
			if err != nil {
				fmt.Println("An Error has occured")
			}
			// reply in the slack channel thru the slacbot the answer to the user's question
			response.Reply(res)

		},
	})

	go printCommandEvents(bot.CommandEvents())

	// This allows for graceful interrupt/cancelation of the server and bots when needed
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := bot.Listen(ctx)

	// If an error has occured, log the error (with Fatal flag)
	if err != nil {
		log.Fatal(err)
	}

}
