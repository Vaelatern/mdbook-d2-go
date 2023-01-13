# mdbook-d2-go

Tiny preprocessor that uses [`d2`](https://github.com/terrastruct/d2) as a library to inline processed SVG in your [mdBook](https://rust-lang.github.io/mdBook/) documents.

## Quickstart

`go build`

then copy the file to your `$PATH` so `mdBook` can find it

Or you can `docker build . -t mdbook-d2-go` and put a `COPY --from=mdbook-d2-go:latest /mdbook-d2-go .` in whatever Dockerfile runs the `mdBook` code.

In your `book.toml`:

```
[preprocessor.d2-go]
layout = "elk"
theme_id = 0
```

### Keeping SVGs skinny

Your SVGs will be full size when loaded in mdbook by default.

Fortunately, a small mdbook theme fixes it:

```
$ cat theme/head.hbs
<style type="text/css">
svg {
        width: inherit;
        height: inherit;
}
</style>
```

## Known bugs

- Currently the SVG generated by `d2` claims the element id `d2-svg` suggesting that two or more of these on the same page may lead to misbehaviours.
- Width may overflow
- The layout engine is not configurable despite most of the plumbing being there

## License

OpenBSD-style ISC
