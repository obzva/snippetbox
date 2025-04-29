# SnippetBox

Snippetbox is a pastebin application made with Go.

This project is a hands-on practice from the book [Let's Go (by Alex Edwards)](https://lets-go.alexedwards.net/).

## Structure

This application follows [this layout](https://go.dev/doc/modules/layout#server-project).

## Little Tweaks

While reading the book, I found a few points that don't fit my opinion or differ from my coding style.

Rather than following it exactly as written, I made a few tweaks.

---

### Application Struct

There are several design decisions I made different from the original version of the book. The most important one is the [`application`](./cmd/web/application.go) struct.

Alex introduces a design where the `application` struct has many methods including route handlers. However, I felt this gave too many responsibilities to one struct, so I decided to make it have only common methods and attributes that could be used throughout the application.

Therefore, the `application` struct was designed to **be passed** as a dependency for many functions or function makers in this project.

---

### Env Variables

Alex introduced the method of setting variables with `flag`, but I prefer environment variables, so I used the `os.Getenv()` method to get the port number and DB connection URI environment variables.

---

### DB

Alex used `MySQL` for the app's database, but I personally wanted to try `PostgreSQL`. Therefore, I chose `PostgreSQL` and [`pgx`](https://github.com/jackc/pgx) for its driver.

---

### Prettified Logger

There was no mention of _prettifying logger_ in the book. However, I found it hard to read logs in plain colors, so I applied the [`tint`](https://github.com/lmittmann/tint) library for the logger.

---

### Tests

I skipped the **testing** chapter since I had already read [this wonderful testing guide](https://quii.gitbook.io/learn-go-with-tests). The main purpose of reading this book was to learn how an experienced Go programmer builds an HTTP web server, not how to test with Go.

---

## Read-more(s)

I want to introduce some topics or keywords that intrigued my curiosity while reading this book.

### Automatic Connection Closing

In the book, Alex says:

> Setting the Connection: Close header on the response acts as a trigger to make Go’s
> HTTP server automatically close the current connection after a response has been sent.

I wanted to find the actual code that does this and [here you are](https://cs.opensource.google/go/go/+/refs/tags/go1.24.2:src/net/http/server.go;l=1397).

```go
if header.get("Connection") == "close" || !keepAlivesEnabled {
	w.closeAfterReply = true
}
```

---

### Internal Packages

Does Go treat directories named `internal` specially?

Well, [Yes](https://go.dev/doc/go1.4#internalpackages).

---

### nil

So, what is `nil`? What is so special about it?

This [two](https://go101.org/article/nil.html) [posts](https://www.dolthub.com/blog/2023-09-08-much-ado-about-nil-things/) explain about that wonderfully.

---

### recover

Why is `recover` only useful inside of deferred functions?

Short explanation:

> **Panic** is a built-in function that stops the ordinary flow of control and begins _panicking_. **When the function F calls panic, execution of F stops, any deferred functions in F are executed normally, and then F returns to its caller.** To the caller, F then behaves like a call to panic. The process continues up the stack until all functions in the current goroutine have returned, at which point the program crashes. Panics can be initiated by invoking panic directly. They can also be caused by runtime errors, such as out-of-bounds array accesses.

Full explanation:

[This link](https://go.dev/blog/defer-panic-and-recover).

And.. a quick example:

[This link](https://gobyexample.com/recover).

---

### Order of Response Writing

Why doesn't Go's net/http package allow `ResponseWriter.WriteHeader()` after `ResponseWriter.Write()`?

It might not be an accurate answer but I could find indirect evidences from [RFC about HTTP](https://www.rfc-editor.org/rfc/rfc9110.html#section-6).

In the third paragraph of the secion 6, it says:

> Framing and control data is sent first, followed by a header section containing fields for the headers table. When a message includes content, the content is sent after the header section, potentially followed by a trailer section that might contain fields for the trailers table.

So I think `net/http` package should be designed to follow this specification.

---

### Server Side Timeout

These [two](https://blog.cloudflare.com/exposing-go-on-the-internet/) [posts](https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/) are awesome primers for this topic if you are new to `server side timeout` concepts.

However, if you feel like these are not enough for you to understand, there are other [blog](https://adam-p.ca/blog/2022/01/golang-http-server-timeouts/) [posts](https://ieftimov.com/posts/make-resilient-golang-net-http-servers-using-timeouts-deadlines-context-cancellation/) explainig about Go `net/http` server's timeout behaviours.
