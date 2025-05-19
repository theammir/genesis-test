<div align="center">
  <h1>⚙️ Genesis Software Engineering School 5.0 Test Case</h1>
  If it ever goes live, API endpoints are accessible <a href="https://server-production-2bf6.up.railway.app/">here</a> (Railway)
  </br>
  <sub>no mailing, though, because deploying mailpit with self-signed certificates on Railway is going to be some extra work</sub>
</div>

## Stack 🚄

* #### The application is written in Go, powered by the [Gin](https://gin-gonic.com/en/) web framework
* #### [Mailpit](https://github.com/axllent/mailpit) as an SMTP server for debugging
* #### SQL migrations with [Goose](https://github.com/pressly/goose/)
* #### The services along with the PostgreSQL DB are brought together by Docker Compose

## Task ⏳

Create and deplot a webservice that fetches current weather forecast from a third-party API and allows subscribing to hourly and daily e-mailing lists for a location.
Use migrations to set up the database.

## Keynotes 🏷️

* API specification was upheld.
* Easy debugging with a separate Compose profile and Mailpit, viewing incoming and outcoming emails locally through a web interface on port 8025.
* Weather data is TTL cached to avoid wasting API calls.
* Email content is rendered from embedded HTML templates.
* TLS is configurable, although no reverse proxy is provided as a service.

## Potential features and nitpicks 🌞

* Simple frontpage to subscribe to events.
* API testing.
* Fixed time for daily forecast updates (a single hourly cron job with a counter is currently used to send out emails, but it helps to simplify the code).

## Setup

Rename `.env.example` to `.env` and set the following variables:
* `HOST` and `PORT` of the web application.
* `WEATHERAPI_KEY` (you can get one [here](https://www.weatherapi.com/my/))
* `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASS`: credentials of the SMTP server. Gmail is a free option.
* `SMTP_FROM` - full email address to send server emails from.

For TLS, specify `TLS_CERT_PATH`, `TLS_KEY_PATH` and optionally `TRUSTED_PLATFORM` (see `.env.example` comment).

Run Docker Compose:
```bash
$ docker compose up
```

For testing, use `./debug.sh up`. It's a shortcut that also creates self-signed TLS certificates for Mailpit, if needed.

## Development thoughts 💡

### Confirmation tokens

There can be multiple confirmation tokens per email, and the spare ones are deleted once the user confirmed the subscription.
However, one token per user is left for unsubscribing, so collisions here are really unwanted.
Some kind of counter, which my dummy implementation was, will most likely be unreliable, and even if not, *I wanted tokens to look random enough for users*.

I went with UUIDs here.
Duplicate UUIDs are still theoretically possible in a large production environment, and it might be worth to look into a Bloom filter.

### Specific Go features

This is actually my very first project in Go. I wanted to find a reason to make one in quite some time.
I liked how easy it was to define structures that can be invalidated based on various conditions, or even values of other fields.

Errors were a tough round: while Go is known for establishing new trends in programming languages, such as supporting documentation out of the box,
I really don't get plain `string` errors, and that they are rarely defined. Not every method specifies *why* it can fail, and the caller (me who's trying to wrap my head around SMTP)
can often only pinpoint the specific issue after logging the error out at runtime.

Maybe I will like them more eventually.

Also, [this](https://github.com/theammir/genesis-test/blob/master/internal/mail/templates.go#L50-L52) is the reason, in my opinion, why formatters exist in the first place -- to handle long lines and let me just write code.
Am I supposed to waste my time indenting this now? Did I have to put thought into it while I was writing this code?
But `gofmt` just refuses to do anything about it.
