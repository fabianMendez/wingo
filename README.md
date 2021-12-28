# Wingo

Service that allows to subscribe for prices updates on wingo's flights

## Serverless Functions

There are three functions on [cmd/functions](cmd/functions) that can be run on [Netlify](https://netlify.com/) or [AWS](https://aws.amazon.com/). These functions
are in charge of managing subscriptions (create, confirm and cancel).

## Main executable

The source code for the main executable is in [cmd/run](cmd/run).

By default it checks for price changes on prices for the (confirmed) subscriptions in a range of 1 month
(you can change the amount of months with the `WINGO_MONTHS` env variable).

It will send emails when:
1. The current price (from the API) is differente from the saved price.
1. There was no saved price but now it's available.
1. There was a saved price but now it's not available.

### Environment variables

|Name|Description|Example|
|---|---|---|
|GH_OWNER|Owner of github repo used to save subscriptions|`user`|
|GH_REPO|Github repo used to save subscriptions|`wingo-data`|
|GH_PATH|Path in github repo used to save subscriptions|`subscriptions`|
|GH_TOKEN|Personal access token used to save subscriptions||
|MG_FROM|Sender to use when sending emails using Mailgun|`User <noreply@user.dev>`|
|MG_API_KEY|API key used to access Mailgun||
|MG_DOMAIN|Domain used to access Mailgun|`mail@user.dev`|


## Future Features
 - [ ] Use a database
 - [ ] Send notifications using WhatsApp
