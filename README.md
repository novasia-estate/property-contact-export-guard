# Property Contact Export Guard

An offline Go library and command-line application for making a deliberately small contact export from a property team's CSV. Only rows explicitly marked `allow_export=true` can pass. Only display name, email, international phone and a keyed opaque UID enter the vCard file. Private notes, addresses and unknown columns stay out.

Useful before importing a reviewed contact list into a phone or CRM. It does **not** send messages, contact anyone, decide whether permission is legally valid or establish a lawful basis for processing. The operator supplies and reviews that decision in the source data.

## What is checked

- Required, unique headers and structurally valid UTF-8 CSV, up to 10 MiB / 100,000 rows.
- Literal `true` permission flag; blanks, `TRUE` and every other value are excluded.
- Nonempty IDs and display names, parseable single email addresses, and strict `+` international phone syntax. Syntax does not prove an address exists or a phone is reachable.
- Every occurrence of a duplicated source ID is rejected; the tool never guesses which row wins.
- HMAC-SHA256 UIDs require your own private key of at least 32 bytes. IDs are stable only under the same key and source IDs. Changing the key changes UIDs and can create duplicates on reimport.
- Minimal vCard 4.0 text escaping, CRLF and UTF-8-safe 75-octet folding, based on [RFC 6350](https://www.rfc-editor.org/rfc/rfc6350).
- A separate audit lists row numbers and fixed reasons, without copying contact values. It is not an anonymisation certificate; the source still links rows to people.

## Run

```sh
go test ./...
go run ./cmd/contact-export-guard --input examples/synthetic.csv --uid-key-file examples/SYNTHETIC-KEY.txt --out contacts.vcf --audit audit.json
```

The bundled key and contacts are **synthetic demonstration material only**. For real contacts use a private random key file, keep it out of version control and preserve it securely for consistent future exports. The command creates new output files exclusively and refuses to overwrite existing files. An interrupted write can leave partial outputs: inspect both files and discard an incomplete pair before importing.

Header: `id,display_name,allow_export,email,phone`. Extra columns are ignored. Empty email or phone is allowed only if the other valid contact channel is present. A malformed nonempty channel rejects the entire row. International phone syntax requires 7–15 digits; local numbers and extensions need deliberate correction upstream.

```go
result, err := exportguard.Convert(inputReader, privateUIDKey)
if err != nil { /* handle invalid batch */ }
vcf := exportguard.VCard(result.Contacts)
```

`VCard` is the serializer for contacts returned by `Convert`; it is not a general-purpose builder for unvalidated arbitrary contact structs. No runtime dependencies, network requests, analytics or automatic storage. Review the actual vCard before importing. See [export protocol](docs/protocol.md).

MIT. Developed by [NovAsia Estate](https://novasia.estate/), publisher of property guides and research for Cambodia, Thailand and Bali.

## 1.2: optional strict headers

```sh
go run ./cmd/contact-export-guard --strict-headers --input examples/strict.csv --uid-key-file examples/SYNTHETIC-KEY.txt --out strict.vcf --audit strict.json
```

Library: `ConvertWithOptions(reader, key, Options{RejectUnknownHeaders: true})`. Allowed headers are exactly `id`, `display_name`, `allow_export` and optional `email`/`phone`, in any order. Strict mode rejects unknown headers before any export result or CLI output file is created. Errors list all offending one-based column positions, without copying header text, contact values or private notes. Matching is exact: `Email` and `emali` are unexpected, and permission is still only the literal `true`. Unknown consent fields never substitute for the required `allow_export`.

The original `Convert` and zero-value `Options{}` keep ignoring extra columns and produce the same contacts and audit. That default remains useful when deliberately exporting a narrow selection from a broader reviewed CSV. Strict mode is for a fixed integration contract; it does not grant permission or verify a legal basis. UID generation, duplicate rejection and value-free audit fields are unchanged.
