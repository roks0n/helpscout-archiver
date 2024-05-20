# Helpscout Archiver

Archive Helpscout Mailbox and Docs to local disk.

This is a very simple command line tool to archive Helpscout Mailbox and Docs to
local disk. If you are looking for a more sophisticated tool, this isn't it.

## About Archiver

### Mailbox

Archives all conversations and puts them into directories based on your mailbox,
folder and conversation. Each conversation is stored as an HTML file and includes
threads and attachments.

### Docs

Archives all docs (no mater of their state) and puts them into directories as
HTML files.

## How to run

```sh
# Archive mailbox
go run cmd/mailbox/mailbox.go <BEARER_TOKEN>
```

```sh
# Archive Docs
go run cmd/docs/docs.go <API_TOKEN>
```

## Authentication tokens

You can get the `BEARER_TOKEN` from the [Helpscout API documentation](https://developer.helpscout.com/mailbox-api/overview/authentication/).

You can get the `API_TOKEN` from the [Helpscout API documentation](https://developer.helpscout.com/docs-api/#your-api-key).
