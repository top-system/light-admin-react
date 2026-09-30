# API response codes

The response shape is unchanged: `code` (string), `message`, optional `data`, and optional `page`. HTTP status communicates the transport category. Clients can use `code` for a stable, more specific decision.

| Code | Meaning |
| --- | --- |
| `00000` | Success |
| `A0400` | Invalid request or validation failure |
| `A0401` | Authentication required or invalid credentials |
| `A0403` | Permission denied |
| `A0404` | Resource not found |
| `A0405` | HTTP method not allowed |
| `A0409` | Unclassified conflict |
| `A0411` | Administrator token expired |
| `A0412` | Member token expired |
| `A0429` | Rate limit exceeded |
| `B1001`–`B1005` | User account and password rules |
| `B1101`–`B1103` | Role rules |
| `B1201`–`B1203` | Menu rules |
| `B1301` | Duplicate configuration key |
| `B1401`–`B1402` | Tenant rules |
| `B1501`–`B1503` | Member rules |
| `B1601` | Incorrect captcha answer |
| `C0500` | Internal server failure |

`errors.BusinessCode` is the registry for exact domain codes. Wrapped sentinel errors are matched with `errors.Is`. Unregistered errors use the HTTP status category; new domain errors should receive a unique code when clients need to distinguish them.
