package exportguard

import (
	"reflect"
	"strings"
	"testing"
)

func TestStrictHeadersRejectAllUnknownPositionsWithoutValues(t *testing.T) {
	s := "id,display_name,allow_export,email,PRIVATE_HEADER,phone,emali\n1,PRIVATE_PERSON,true,e@example.invalid,SECRET,,PRIVATE_EMAIL\n"
	r, err := ConvertWithOptions(strings.NewReader(s), testKey, Options{RejectUnknownHeaders: true})
	if err == nil || err.Error() != "unexpected CSV header columns: 5,7" || !reflect.DeepEqual(r, Result{}) {
		t.Fatal(r, err)
	}
}

func TestZeroOptionsPreserveLegacyResultExactly(t *testing.T) {
	s := csvData("1,Example,true,e@example.invalid,,SECRET")
	a, e := Convert(strings.NewReader(s), testKey)
	b, f := ConvertWithOptions(strings.NewReader(s), testKey, Options{})
	if e != nil || f != nil || !reflect.DeepEqual(a, b) {
		t.Fatal(a, b, e, f)
	}
}

func TestStrictHeadersRetainConsentDuplicateAndUIDRules(t *testing.T) {
	s := "phone,email,allow_export,id,display_name\n,e@example.invalid,true,1,Example\n"
	a, e := ConvertWithOptions(strings.NewReader(s), testKey, Options{RejectUnknownHeaders: true})
	b, f := Convert(strings.NewReader(s), testKey)
	if e != nil || f != nil || !reflect.DeepEqual(a, b) || len(a.Contacts) != 1 {
		t.Fatal(a, b, e, f)
	}
	for _, row := range []string{"1,A,false,a@example.invalid", "1,A,TRUE,a@example.invalid"} {
		r, e := ConvertWithOptions(strings.NewReader("id,display_name,allow_export,email\n"+row+"\n"), testKey, Options{RejectUnknownHeaders: true})
		if e != nil || r.Audit.Issues[0].Reason != "export_not_explicitly_allowed" {
			t.Fatal(r, e)
		}
	}
	r, e := ConvertWithOptions(strings.NewReader("id,display_name,allow_export,email\n1,A,true,a@example.invalid\n1,B,true,b@example.invalid\n"), testKey, Options{RejectUnknownHeaders: true})
	if e != nil || len(r.Contacts) != 0 || len(r.Audit.Issues) != 2 {
		t.Fatal(r, e)
	}
}

func TestStrictHeaderOnlyAndMalformedInputs(t *testing.T) {
	r, e := ConvertWithOptions(strings.NewReader("id,display_name,allow_export\n"), testKey, Options{RejectUnknownHeaders: true})
	if e != nil || r.Audit.Rows != 0 {
		t.Fatal(r, e)
	}
	for _, s := range []string{"id,display_name,allow_export,Email\n", "id,display_name,allow_export,\n", "id,display_name,allow_export,id\n", "id,display_name,allow_exprot\n"} {
		if _, e := ConvertWithOptions(strings.NewReader(s), testKey, Options{RejectUnknownHeaders: true}); e == nil {
			t.Fatal("unexpected headers accepted")
		}
	}
}
