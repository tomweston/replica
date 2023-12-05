<div align="center">

<img src="./assets/logo-512px.png" alt="Replica Logo" width="100">

# REPLICΛ

[Overview](#overview) •
[Key Features](#key-features) •
[Prerequisites](#prerequisites) •
[Permissions and Settings](#permissions-and-settings) •
[Contributing](#-contributing) •
[License](#-license)

`Replica` Is a Slack app that integrates with Datadog, allowing users to seamlessly create replicas of dashboards directly from Slack.

</div>

## Overview

The app provides several functionalities, including:

1. **Event Handling**: Utilizing Slack's Events API and Socket Mode, the app listens for specific Slack events.
2. **Dashboard Replication**: It fetches available Datadog dashboards and presents a dropdown list to the user. After a dashboard selection, it creates a replica in Datadog.
3. **Name Generation**: Generates a fun, randomized name for the replica dashboard.
4. **Environment Configuration**: The app sources its configuration from an `.env` file.

## Key Features

- **Slack Interactivity**:
  - Replies with a greeting to "hello" messages.
  - Opens a dashboard selection modal on the 'replica' shortcut command.
  - Posts messages with links to the replicated dashboard and a merge option.

- **Datadog Integration**:
  - Retrieves available dashboards from Datadog.
  - Facilitates dashboard replication.

## Prerequisites

To set up the app:

1. Clone this repository.
2. Copy `.env.example` and rename it to `.env`.
3. Fill in the required values in `.env`:

    ```sh
    SLACK_APP_TOKEN=xapp-xxxxxx
    SLACK_BOT_TOKEN=xoxb-xxxxxx
    DATADOG_API_KEY=
    DATADOG_APP_KEY=
    SLACK_CHANNEL_ID=
    ```

4. Install necessary Go libraries with `go get`.
5. Run the app using `go run main.go`.

## Permissions and Settings

### OAuth Scopes

The `Replica` app requires the following OAuth scopes for functionality:

#### User Scopes

- `chat:write`: To send messages as the user.

#### Bot Scopes

- `chat:write`: To send messages in channels.
- `commands`: To add slash commands and shortcuts.
- `app_mentions:read`: To read messages that mention the app.
- `channels:history`: To access the message history of channels.
- `channels:read`: To view channels in Slack.
- `im:read`: To view direct messages.
- `im:write`: To send direct messages.
- `mpim:history`: To access multi-party direct message history.
- `im:history`: To access direct message history.
- `groups:history`: To access private channel message history.

### Event Subscriptions

The app listens to the following bot events:

- `app_mention`: When the app is mentioned.
- `message.channels`: Messages in public channels.
- `message.groups`: Messages in private channels.
- `message.im`: Direct messages.
- `message.mpim`: Multi-party direct messages.

### Interactivity and Shortcuts

- Interactivity is enabled for the app.
- The app includes a global shortcut named "Create a Replica" for creating Datadog dashboard replicas.
- Request URLs for interactivity and message menu options are set to `https://localhost:8080`.

### Additional Settings

- Socket mode is enabled, allowing the app to use WebSockets for receiving events.
- Organization-wide deployment is not enabled (`org_deploy_enabled: false`).
- Token rotation is not enabled (`token_rotation_enabled: false`).

## 🤝 Contributing

Contributions, issues and feature requests are welcome!

## 📄 License

This project is [MIT](./LICENSE) licensed.

## Author

[![LinkedIn](https://img.shields.io/badge/linkedin-%230077B5.svg?&style=for-the-badge&logo=linkedin&logoColor=white)](https://www.linkedin.com/in/westontom)
[![Twitter](https://img.shields.io/badge/@tomweston-%231DA1F2.svg?&style=for-the-badge&logo=x&logoColor=white)](https://twitter.com/tomweston)