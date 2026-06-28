## Intent Diff: Grade C — Material omissions
**Confidence**: high | **Score**: 0.45 | **Highest risk**: contract

### Attention Map
| Priority | File | Reason |
|----------|------|--------|
| high | `auth/session.go` | Session expiry logic changed |
| medium | `api/handler.go` | New error code added |

### Mismatches (1)
**[contract]** Auth session expiry changed — severity: high
> Session timeout reduced from 24h to 1h

<details><summary>Full analysis</summary>

**Evidence** (Auth session expiry changed): auth/session.go

</details>
