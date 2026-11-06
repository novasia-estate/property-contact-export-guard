package exportguard

import "strings"

func escape(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	return strings.ReplaceAll(s, ",", `\,`)
}

// fold splits content lines at at most 75 UTF-8 octets, without splitting a code point.
// Continuations start with one space, counted in that line's 75-byte limit.
func fold(line string) string {
	var out strings.Builder
	count := 0
	for _, r := range line {
		s := string(r)
		if count+len(s) > 75 {
			out.WriteString("\r\n ")
			count = 1
		}
		out.WriteString(s)
		count += len(s)
	}
	out.WriteString("\r\n")
	return out.String()
}

// VCard serializes contacts returned by Convert to a minimal vCard 4.0 profile.
// It is not a general-purpose vCard generator. Revalidate arbitrary Contact values
// through Convert before calling it. Email/telephone values are supplied unmodified.
func VCard(contacts []Contact) string {
	var out strings.Builder
	for _, c := range contacts {
		for _, line := range []string{"BEGIN:VCARD", "VERSION:4.0", "UID;VALUE=text:" + c.UID, "FN:" + escape(c.DisplayName)} {
			out.WriteString(fold(line))
		}
		if c.Email != "" {
			out.WriteString(fold("EMAIL:" + escape(c.Email)))
		}
		if c.Phone != "" {
			out.WriteString(fold("TEL;VALUE=uri:tel:" + c.Phone))
		}
		out.WriteString("END:VCARD\r\n")
	}
	return out.String()
}
