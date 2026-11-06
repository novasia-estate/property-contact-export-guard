package exportguard

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"io"
	"net/mail"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Contact contains only the fields allowed into an export. Source IDs and notes are absent.
type Contact struct{ UID, DisplayName, Email, Phone string }

// Issue identifies an input row and a fixed reason, without copying contact values.
type Issue struct {
	Row    int    `json:"row"`
	Reason string `json:"reason"`
}

// Audit is a value-free review of inclusion decisions. Row numbers start at 2 after the header.
type Audit struct {
	Rows     int     `json:"rows"`
	Exported int     `json:"exported"`
	Issues   []Issue `json:"issues"`
}

// Result contains permitted contacts and an audit. Keep contacts private even though the audit omits their values.
type Result struct {
	Contacts []Contact
	Audit    Audit
}

var phonePattern = regexp.MustCompile(`^\+[1-9][0-9]{6,14}$`)

// Convert reads a UTF-8 CSV with id, display_name, allow_export and optional email/phone.
// Only the literal allow_export value "true" permits export. Duplicate source IDs
// are all rejected. A separate 32-byte or longer HMAC key creates stable opaque UIDs.
// Maximum input size is 10 MiB and maximum row count is 100000. Unexpected columns
// are ignored; no addresses, notes, organisations or custom fields are exported.
func Convert(input io.Reader, uidKey []byte) (Result, error) {
	if len(uidKey) < 32 {
		return Result{}, errors.New("UID key requires at least 32 bytes")
	}
	data, err := io.ReadAll(io.LimitReader(input, 10*1024*1024+1))
	if err != nil {
		return Result{}, err
	}
	if len(data) > 10*1024*1024 {
		return Result{}, errors.New("input exceeds 10 MiB")
	}
	if !utf8.Valid(data) {
		return Result{}, errors.New("input is not UTF-8")
	}
	rows, err := csv.NewReader(strings.NewReader(string(data))).ReadAll()
	if err != nil {
		return Result{}, errors.New("invalid CSV structure")
	}
	if len(rows) == 0 {
		return Result{}, errors.New("CSV header is required")
	}
	if len(rows) > 100001 {
		return Result{}, errors.New("maximum 100000 contact rows")
	}
	columns := map[string]int{}
	for i, s := range rows[0] {
		if _, ok := columns[s]; ok {
			return Result{}, errors.New("duplicate CSV header")
		}
		columns[s] = i
	}
	for _, s := range []string{"id", "display_name", "allow_export"} {
		if _, ok := columns[s]; !ok {
			return Result{}, errors.New("missing required CSV header: " + s)
		}
	}
	get := func(r []string, k string) string {
		if i, ok := columns[k]; ok {
			return r[i]
		}
		return ""
	}
	counts := map[string]int{}
	for _, r := range rows[1:] {
		counts[get(r, "id")]++
	}
	result := Result{Audit: Audit{Rows: len(rows) - 1, Issues: []Issue{}}, Contacts: []Contact{}}
	for n, r := range rows[1:] {
		reject := func(reason string) {
			result.Audit.Issues = append(result.Audit.Issues, Issue{Row: n + 2, Reason: reason})
		}
		if get(r, "allow_export") != "true" {
			reject("export_not_explicitly_allowed")
			continue
		}
		id, name, email, phone := get(r, "id"), get(r, "display_name"), get(r, "email"), get(r, "phone")
		if strings.TrimSpace(id) == "" || strings.IndexFunc(id, unicode.IsControl) >= 0 {
			reject("invalid_source_id")
			continue
		}
		if counts[id] > 1 {
			reject("duplicate_source_id")
			continue
		}
		if strings.TrimSpace(name) == "" || len(name) > 2000 || strings.IndexFunc(name, func(r rune) bool { return unicode.IsControl(r) && r != '\n' && r != '\r' }) >= 0 {
			reject("invalid_display_name")
			continue
		}
		if email != "" {
			a, e := mail.ParseAddress(email)
			if e != nil || a.Address != email || strings.ContainsAny(email, "\r\n") {
				reject("invalid_email")
				continue
			}
		}
		if phone != "" && !phonePattern.MatchString(phone) {
			reject("invalid_international_phone")
			continue
		}
		if email == "" && phone == "" {
			reject("no_contact_channel")
			continue
		}
		mac := hmac.New(sha256.New, uidKey)
		mac.Write([]byte("contact-export-v1\x00" + id))
		result.Contacts = append(result.Contacts, Contact{UID: "cg1_" + hex.EncodeToString(mac.Sum(nil)), DisplayName: name, Email: email, Phone: phone})
	}
	result.Audit.Exported = len(result.Contacts)
	return result, nil
}
