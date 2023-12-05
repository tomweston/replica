package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

type DatadogDashboard struct {
	ID    string
	Title string
}

type Payload struct {
	View struct {
		State struct {
			Values map[string]map[string]struct {
				SelectedOption struct {
					Value string `json:"value"`
				} `json:"selected_option"`
			} `json:"values"`
		} `json:"state"`
	} `json:"view"`
}

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

func replicaName() string {
	adjectives := []string{
		"happy", "elated", "sad", "angry", "furious", "mysterious", "bright", "dark", "silent", "loud",
		"luminous", "calm", "serene", "fluffy", "spiky", "colorful", "vibrant", "gloomy", "slimy", "grumpy",
		"joyful", "optimistic", "pessimistic", "melodic", "harsh", "hollow", "stuffed", "bulky", "slender", "brave",
		"meek", "heroic", "cowardly", "glittering", "dull", "shiny", "matte", "spherical", "flat", "crispy",
		"soft", "rigid", "flexible", "sturdy", "flimsy", "chunky", "sparse", "dense", "witty", "dim",
		"boisterous", "muted", "candid", "staged", "authentic", "forged", "moving", "still", "animated", "lifelike",
		"distant", "nearby", "exotic", "common", "splendid", "dreary", "beaming", "sour", "spicy", "mild",
		"scalding", "icy", "steamy", "frozen", "thunderous", "silent", "noisy", "hushed", "rough", "smooth",
		"plush", "wrinkled", "muddy", "clean", "filthy", "spotless", "ragged", "pristine", "aged", "new",
	}

	verbs := []string{
		"run", "jump", "swim", "dive", "climb", "crawl", "sing", "shout", "whisper", "write",
		"sketch", "draw", "paint", "build", "destroy", "dance", "laugh", "cry", "sulk", "ponder",
		"wonder", "dream", "hope", "fear", "create", "invent", "discover", "explore", "wander", "stumble",
		"grasp", "clutch", "release", "catch", "throw", "punch", "kick", "push", "pull", "lift",
		"drop", "break", "fix", "mend", "weld", "carve", "sculpt", "measure", "design", "plot",
		"scheme", "act", "perform", "entertain", "calculate", "ponder", "think", "believe", "doubt", "guess",
		"play", "work", "rest", "sleep", "awaken", "startle", "surprise", "frighten", "scare", "console",
		"comfort", "coax", "convince", "persuade", "dissuade", "begin", "end", "commence", "terminate", "introduce",
		"eliminate", "increase", "decrease", "inflate", "deflate", "expand", "contract", "magnify", "diminish", "accelerate",
	}

	randomAdjective := adjectives[r.Intn(len(adjectives))]
	randomVerb := verbs[r.Intn(len(verbs))]

	return fmt.Sprintf("%s-%s", randomAdjective, randomVerb)
}

func CreateDatadogContext() (context.Context, *datadog.Configuration) {
	apiKey := datadog.APIKey{Key: os.Getenv("DATADOG_API_KEY")}
	appKey := datadog.APIKey{Key: os.Getenv("DATADOG_APP_KEY")}

	ctx := context.WithValue(
		context.Background(),
		datadog.ContextAPIKeys,
		map[string]datadog.APIKey{
			"apiKeyAuth": apiKey,
			"appKeyAuth": appKey,
		},
	)

	configuration := datadog.NewConfiguration()
	configuration.Host = "api.datadoghq.eu"

	return ctx, configuration
}

func FetchDatadogDashboards() ([]DatadogDashboard, error) {

	ctx, configuration := CreateDatadogContext()
	apiClient := datadog.NewAPIClient(configuration)
	api := datadogV1.NewDashboardsApi(apiClient)

	resp, _, err := api.ListDashboards(ctx, *datadogV1.NewListDashboardsOptionalParameters().WithFilterShared(false))
	if err != nil {
		return nil, err
	}

	// Slack limits the number of options in a dropdown to 100 :( (https://api.slack.com/reference/block-kit/block-elements#static_select
	// and https://api.slack.com/reference/block-kit/composition-objects#option)
	// So we'll just return the first 99 dashboards for now until we can figure out a better way to handle this. ¯\_(ツ)_/¯
	// Options:
	// 1. We could return dashboards that the belong to the slack caller's team. This would be ideal but would require retrieving the slack caller's team ID and then matching that to the team ID in the dashboard metadata. This would require a lot of API calls and would be slow.
	// 2. We could return dashboards that have a tag of "prod" or similar. This would be a quick fix but not ideal as they could still exceed the 100 option limit. Also, the tool shouldideally not be limited in this way.
	var dashboards []DatadogDashboard
	for _, dashboard := range resp.Dashboards {
		if len(dashboards) >= 99 { // 99 is the max number of options allowed in a dropdown
			break
		}
		dashboards = append(dashboards, DatadogDashboard{
			ID:    dashboard.GetId(),
			Title: dashboard.GetTitle(),
		})
	}

	return dashboards, nil
}

// handleViewSubmission should handle all of the logic for when a user submits a view. However removing this function order seems to delay the Ack response from the socketmode server and causes the modal to not close. So for now, we'll just leave this here and not use it. :)
// TODO: Figure out why it's timing out..
func handleViewSubmission(payload slack.InteractionCallback) {
	selectedOption := payload.View.State.Values["dropdown_block_id"]["dropdown_action_id"].SelectedOption.Value
	log.Printf("Selected Option: %v\n", selectedOption)
}

func CloneDashboardAndReturnReplicaLink(selectedDashboardID, replicaName string) (string, error) {

	ctx, configuration := CreateDatadogContext()
	apiClient := datadog.NewAPIClient(configuration)
	api := datadogV1.NewDashboardsApi(apiClient)

	dashboardDetails, _, err := api.GetDashboard(ctx, selectedDashboardID)
	if err != nil {
		return "", err
	}

	dashboardDetails.SetTitle(replicaName)

	replica, _, err := api.CreateDashboard(ctx, dashboardDetails)
	if err != nil {
		return "", err
	}
	baseURL := "https://app.datadoghq.eu/dashboard/"
	return baseURL + replica.GetId(), nil
}

func openReplicaModal(api *slack.Client, triggerID string) {
	dashboards, err := FetchDatadogDashboards()
	if err != nil {
		log.Printf("Failed to fetch dashboards: %v", err)
		return
	}

	options := make([]*slack.OptionBlockObject, len(dashboards))
	for i, dashboard := range dashboards {
		optionText := slack.NewTextBlockObject("plain_text", dashboard.Title, false, false)
		options[i] = slack.NewOptionBlockObject(dashboard.ID, optionText, optionText)
	}

	dropdown := slack.NewOptionsSelectBlockElement("static_select", nil, "dropdown_action_id", options...)

	inputBlockWithDropdown := slack.NewInputBlock(
		"dropdown_block_id",
		slack.NewTextBlockObject("plain_text", "Select a Dashboard", false, false),
		slack.NewTextBlockObject("plain_text", "Ensure you select the correct dashboard from the list", false, false),
		dropdown,
	)

	descriptionText := "Please choose a Datadog dashboard from the dropdown below that you wish to replicate. Once you've made your selection, click 'Create' to generate a unique replica link."
	descriptionBlock := slack.NewSectionBlock(
		slack.NewTextBlockObject("mrkdwn", descriptionText, false, false),
		nil, nil,
	)

	CreateReplicaModalView := slack.ModalViewRequest{
		Type:       "modal",
		CallbackID: "modal-id",
		Title: slack.NewTextBlockObject(
			"plain_text",
			"Create a Replica",
			false,
			false,
		),
		Submit: slack.NewTextBlockObject(
			"plain_text",
			"Create",
			false,
			false,
		),
		Close: slack.NewTextBlockObject(
			"plain_text",
			"Cancel",
			false,
			false,
		),
		Blocks: slack.Blocks{
			BlockSet: []slack.Block{
				descriptionBlock,
				inputBlockWithDropdown,
			},
		},
	}

	_, err = api.OpenView(triggerID, CreateReplicaModalView)
	if err != nil {
		log.Printf("Failed to open a modal: %v", err)
	}
}

func main() {

	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}

	requiredEnvVars := []string{"DATADOG_API_KEY", "DATADOG_APP_KEY", "SLACK_BOT_TOKEN", "SLACK_APP_TOKEN", "SLACK_CHANNEL_ID"}
	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			fmt.Printf("Error: %s not set in .env\n", envVar)
			return
		}
	}
	log.Println("Environment variables successfully read from .env")

	webApi := slack.New(
		os.Getenv("SLACK_BOT_TOKEN"),
		slack.OptionAppLevelToken(os.Getenv("SLACK_APP_TOKEN")),
		slack.OptionDebug(true),
		slack.OptionLog(log.New(os.Stdout, "api: ", log.Lshortfile|log.LstdFlags)),
	)
	socketMode := socketmode.New(
		webApi,
		socketmode.OptionDebug(true),
		socketmode.OptionLog(log.New(os.Stdout, "sm: ", log.Lshortfile|log.LstdFlags)),
	)
	authTest, authTestErr := webApi.AuthTest()
	if authTestErr != nil {
		fmt.Fprintf(os.Stderr, "SLACK_BOT_TOKEN is invalid: %v\n", authTestErr)
		os.Exit(1)
	}
	selfUserId := authTest.UserID

	channelID := os.Getenv("SLACK_CHANNEL_ID")

	go func() {
		for envelope := range socketMode.Events {
			switch envelope.Type {

			case socketmode.EventTypeSlashCommand:
				cmd, ok := envelope.Data.(slack.SlashCommand)
				if !ok {
					log.Printf("Ignored slash command: %v", envelope.Data)
					continue
				}

				if cmd.Command == "/rep" {
					socketMode.Ack(*envelope.Request)
					openReplicaModal(webApi, cmd.TriggerID)
				}

			case socketmode.EventTypeEventsAPI:
				socketMode.Ack(*envelope.Request)
				eventPayload, _ := envelope.Data.(slackevents.EventsAPIEvent)
				switch eventPayload.Type {
				case slackevents.CallbackEvent:
					switch event := eventPayload.InnerEvent.Data.(type) {
					case *slackevents.MessageEvent:
						if event.User != selfUserId &&
							strings.Contains(strings.ToLower(event.Text), "hello") {
							_, _, err := webApi.PostMessage(
								event.Channel,
								slack.MsgOptionText(
									fmt.Sprintf(":wave: Hi there, <@%v>!", event.User),
									false,
								),
							)
							if err != nil {
								log.Printf("Failed to reply: %v", err)
							}
						}
					default:
						socketMode.Debugf("Skipped: %v", event)
					}
				default:
					socketMode.Debugf("unsupported Events API eventPayload received")
				}
			case socketmode.EventTypeInteractive:

				payload, _ := envelope.Data.(slack.InteractionCallback)
				switch payload.Type {
				case slack.InteractionTypeShortcut:
					if payload.CallbackID == "replica" {
						socketMode.Ack(*envelope.Request)

						dashboards, err := FetchDatadogDashboards()
						if err != nil {
							log.Printf("Failed to fetch dashboards: %v", err)
							return
						}

						options := make([]*slack.OptionBlockObject, len(dashboards))
						for i, dashboard := range dashboards {
							optionText := slack.NewTextBlockObject("plain_text", dashboard.Title, false, false)
							options[i] = slack.NewOptionBlockObject(dashboard.ID, optionText, optionText)
						}

						dropdown := slack.NewOptionsSelectBlockElement("static_select", nil, "dropdown_action_id", options...)

						inputBlockWithDropdown := slack.NewInputBlock(
							"dropdown_block_id",
							slack.NewTextBlockObject("plain_text", "Select a Dashboard", false, false),
							slack.NewTextBlockObject("plain_text", "Ensure you select the correct dashboard from the list", false, false),
							dropdown,
						)

						descriptionText := "Please choose a Datadog dashboard from the dropdown below that you wish to relpicate. Once you've made your selection, click 'Create' to generate a unique replica link."
						descriptionBlock := slack.NewSectionBlock(
							slack.NewTextBlockObject("mrkdwn", descriptionText, false, false),
							nil, nil,
						)

						CreateReplicaModalView := slack.ModalViewRequest{
							Type:       "modal",
							CallbackID: "modal-id",
							Title: slack.NewTextBlockObject(
								"plain_text",
								"Create a Replica",
								false,
								false,
							),
							Submit: slack.NewTextBlockObject(
								"plain_text",
								"Create",
								false,
								false,
							),
							Close: slack.NewTextBlockObject(
								"plain_text",
								"Cancel",
								false,
								false,
							),
							Blocks: slack.Blocks{
								BlockSet: []slack.Block{
									descriptionBlock,
									inputBlockWithDropdown,
								},
							},
						}
						resp, err := webApi.OpenView(payload.TriggerID, CreateReplicaModalView)
						if err != nil {
							log.Printf("Failed to open a modal: %v", err)
						}
						socketMode.Debugf("views.open response: %v", resp)
					}
				case slack.InteractionTypeViewSubmission:
					// socketMode.Ack(*envelope.Request) // Acknowledge the receipt of the payload right away
					if payload.CallbackID == "modal-id" {
						handleViewSubmission(payload)
					}
					socketMode.Ack(*envelope.Request)
					initiatorUserID := payload.User.ID
					selectedDashboardID := payload.View.State.Values["dropdown_block_id"]["dropdown_action_id"].SelectedOption.Value
					selectedDashboardName := payload.View.State.Values["dropdown_block_id"]["dropdown_action_id"].SelectedOption.Text.Text

					replicaName := replicaName()
					replicaLink, err := CloneDashboardAndReturnReplicaLink(selectedDashboardID, replicaName)
					if err != nil {
						log.Printf("Failed to clone dashboard: %v", err)
						return
					}

					dashboardText := slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("<@%s>\n\nSelected dashboard: *%s*", initiatorUserID, selectedDashboardName), false, false)
					dashboardBlock := slack.NewSectionBlock(dashboardText, nil, nil)

					replicaText := slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("Generated replica name: *%s*", replicaName), false, false)
					replicaBlock := slack.NewSectionBlock(replicaText, nil, nil)

					dashboardButtonText := slack.NewTextBlockObject("plain_text", "Open Replica", false, false)
					dashboardButton := slack.NewButtonBlockElement("", "dashboard", dashboardButtonText)
					dashboardButton.URL = replicaLink

					mergeButtonText := slack.NewTextBlockObject("plain_text", "Merge Changes", false, false)
					mergeButton := slack.NewButtonBlockElement("", "merge", mergeButtonText)
					mergeButton.URL = "https://github.com/tomweston/replica/pulls"

					buttons := slack.NewActionBlock("", dashboardButton, mergeButton)

					msgResponse, _, err := webApi.PostMessage(
						channelID,
						slack.MsgOptionBlocks(dashboardBlock, replicaBlock, buttons),
					)

					if err != nil {
						log.Printf("Failed to send message: %v", err)
					} else {
						log.Printf("Message response: %+v", msgResponse)
					}

					// socketMode.Debugf("Submitted Data: %v", payload.View.State.Values)
					socketMode.Ack(*envelope.Request)

				default:
					// socketMode.Debugf("Skipped: %v", payload)
				}

			default:
				// socketMode.Debugf("Skipped: %v", envelope.Type)
			}
		}
	}()

	socketMode.Run()
}
