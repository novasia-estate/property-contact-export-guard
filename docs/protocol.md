# Review an export

1. Work on a deliberately prepared CSV and confirm the intended destination and permission in your own process.
2. Keep a private random UID key stable across repeated exports from that source namespace. Use separate keys for unrelated source systems unless their IDs are explicitly coordinated.
3. Run the tool into new output paths. Inspect the audit and correct excluded rows upstream; do not force a true flag to suppress a rejection.
4. Inspect the actual names and channels in the vCard. Import only after this review.
5. Keep source and exports under your organisation's access rules. The audit omits values but row numbers can still be linked to the source.

The tool intentionally does not export organisation, address, notes, property interests, lead scores or inferred personal categories. It is a minimal contact-transfer profile, not a full CRM migration format. HMAC UIDs conceal raw source IDs from the vCard but the vCard still contains personal contact details.
