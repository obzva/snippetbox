# SnippetBox

Snippetbox is a pastebin application made with Go.

This project is a hands-on practice from the book [Let's Go (by Alex Edwards)](https://lets-go.alexedwards.net/).

## Structure

This application follows [this layout](https://go.dev/doc/modules/layout#server-project).

## Little Tweaks

While reading the book, I found a few points that don't fit my opinion or differ from my coding style.

Rather than following it exactly as written, I made a few tweaks.

### Application Struct

There are several design decisions I made different from the original version of the book. The most important one is the [`application`](./cmd/web/application.go) struct.

Alex introduces a design where the `application` struct has many methods including route handlers. However, I felt this gave too many responsibilities to one struct, so I decided to make it have only common methods and attributes that could be used throughout the application.

Therefore, the `application` struct was designed to **be passed** as a dependency for many functions or function makers in this project.

### Env Variables

Alex introduced the method of setting variables with `flag`, but I prefer environment variables, so I used the `os.Getenv()` method to get the port number and DB connection URI environment variables.

### DB

Alex used `MySQL` for the app's database, but I personally wanted to try `PostgreSQL`. Therefore, I chose `PostgreSQL` and [`pgx`](https://github.com/jackc/pgx) for its driver.

### Prettified Logger

There was no mention of _prettifying logger_ in the book. However, I found it hard to read logs in plain colors, so I applied the [`tint`](https://github.com/lmittmann/tint) library for the logger.

### Tests

I skipped the **testing** chapter since I had already read [this wonderful testing guide](https://quii.gitbook.io/learn-go-with-tests). The main purpose of reading this book was to learn how an experienced Go programmer builds an HTTP web server, not how to test with Go.
