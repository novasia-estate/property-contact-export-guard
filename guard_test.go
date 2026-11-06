package exportguard

import (
	"encoding/json"
	"strings"
	"testing"
	"unicode/utf8"
)

var testKey = []byte("SYNTHETIC-TEST-KEY-32-BYTES-ONLY!!")

func csvData(row string) string {
	return "id,display_name,allow_export,email,phone,private_notes\n" + row + "\n"
}
func TestAcceptedMinimal(t *testing.T) {
	r, e := Convert(strings.NewReader(csvData("1,Example Editor,true,editor@example.invalid,+85512345678,SECRET")), testKey)
	if e != nil || len(r.Contacts) != 1 {
		t.Fatal(r, e)
	}
	s := VCard(r.Contacts)
	if strings.Contains(s, "SECRET") || !strings.Contains(s, "VERSION:4.0\r\n") {
		t.Fatal(s)
	}
}
func TestRejects(t *testing.T) {
	for _, tc := range []struct{ row, reason string }{
		{"1,Example,false,e@example.invalid,,secret", "export_not_explicitly_allowed"},
		{"1,Example,TRUE,e@example.invalid,,secret", "export_not_explicitly_allowed"},
		{"1,Example,,e@example.invalid,,secret", "export_not_explicitly_allowed"},
		{"1,,true,e@example.invalid,,secret", "invalid_display_name"},
		{"1,Example,true,broken,,secret", "invalid_email"},
		{"1,Example,true,,01234,secret", "invalid_international_phone"},
		{"1,Example,true,,,secret", "no_contact_channel"},
		{",Example,true,e@example.invalid,,secret", "invalid_source_id"},
	} {
		t.Run(tc.reason+tc.row, func(t *testing.T) {
			r, e := Convert(strings.NewReader(csvData(tc.row)), testKey)
			if e != nil || len(r.Contacts) != 0 || r.Audit.Issues[0].Reason != tc.reason {
				t.Fatal(r, e)
			}
		})
	}
}
func TestDuplicateIDsRejectAll(t *testing.T) {
	s := csvData("1,A,true,a@example.invalid,,x") + "1,B,true,b@example.invalid,,y\n"
	r, e := Convert(strings.NewReader(s), testKey)
	if e != nil || len(r.Contacts) != 0 || len(r.Audit.Issues) != 2 {
		t.Fatal(r, e)
	}
}
func TestUIDKeyIsolation(t *testing.T) {
	s := csvData("1,A,true,a@example.invalid,,x")
	a, _ := Convert(strings.NewReader(s), testKey)
	b, _ := Convert(strings.NewReader(s), []byte(strings.Repeat("b", 32)))
	c, _ := Convert(strings.NewReader(s), testKey)
	if a.Contacts[0].UID == b.Contacts[0].UID || a.Contacts[0].UID != c.Contacts[0].UID {
		t.Fatal("unstable or unkeyed UID")
	}
}
func TestAuditDoesNotCopyValues(t *testing.T) {
	r, _ := Convert(strings.NewReader(csvData("PRIVATEID,PRIVATEPERSON,true,BAD_EMAIL,,SECRET")), testKey)
	b, _ := json.Marshal(r.Audit)
	for _, s := range []string{"PRIVATEID", "PRIVATEPERSON", "BAD_EMAIL", "SECRET"} {
		if strings.Contains(string(b), s) {
			t.Fatal(string(b))
		}
	}
}
func TestControlsAndEscaping(t *testing.T) {
	r, e := Convert(strings.NewReader(csvData("1,\"Example\nEND:VCARD\nBEGIN:VCARD\",true,e@example.invalid,,x")), testKey)
	if e != nil {
		t.Fatal(e)
	}
	s := VCard(r.Contacts)
	if strings.Count(s, "\r\nEND:VCARD\r\n") != 1 || strings.Contains(s, "\r\nBEGIN:VCARD\r\n") {
		t.Fatal(s)
	}
}
func TestUTF8Folding(t *testing.T) {
	line := "FN:" + strings.Repeat("Ж🙂", 70)
	s := fold(line)
	for _, l := range strings.Split(strings.TrimSuffix(s, "\r\n"), "\r\n") {
		if len(l) > 75 || !utf8.ValidString(l) {
			t.Fatal(l)
		}
	}
	if strings.ReplaceAll(strings.TrimSuffix(s, "\r\n"), "\r\n ", "") != line {
		t.Fatal("unfold mismatch")
	}
}
func TestMalformed(t *testing.T) {
	for _, s := range []string{"", "id,id,display_name,allow_export\n", "id,display_name\n", "id,display_name,allow_export\n1,A\n", string([]byte{0xff})} {
		if _, e := Convert(strings.NewReader(s), testKey); e == nil {
			t.Fatal("accepted malformed CSV", s)
		}
	}
}
func TestShortKey(t *testing.T) {
	if _, e := Convert(strings.NewReader(csvData("1,A,true,e@example.invalid,,x")), []byte("short")); e == nil {
		t.Fatal("short key")
	}
}
func TestHeaderOnly(t *testing.T) {
	r, e := Convert(strings.NewReader("id,display_name,allow_export\n"), testKey)
	if e != nil || r.Audit.Rows != 0 || VCard(r.Contacts) != "" {
		t.Fatal(r, e)
	}
}
