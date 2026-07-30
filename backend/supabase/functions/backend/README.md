# Backend Edge Function

The `backend` function exposes four application operations:

| Method | Route | Authentication | Purpose |
| --- | --- | --- | --- |
| `POST` | `/register` | Publishable API key | Get or create the access code for a device |
| `POST` | `/login` | Publishable API key | Exchange an access code for a Supabase session |
| `POST` | `/score` | Supabase user JWT | Create or replace the current user's score |
| `GET` | `/scores` | Publishable API key | Read the global leaderboard |

`/access-code` aliases `/register`. The singular and plural score routes are
also accepted for both score operations.

## Configuration

Set one stable secret with at least 32 random characters:

```text
ACCESS_CODE_PEPPER=...
```

Do not rotate this value without migrating every account. It is used to derive
device hashes, access codes, lookup hashes, and internal Supabase Auth
passwords.

Email/password authentication must be enabled in Supabase Auth. The generated
email addresses are internal identifiers; they are confirmed on creation and
are never returned by the API.

Apply and deploy:

```sh
supabase db push
supabase secrets set ACCESS_CODE_PEPPER=...
supabase functions deploy backend
```

The function has `verify_jwt = false` because registration and login occur
before a user JWT exists. `@supabase/server` validates the publishable API key
or user JWT inside the function.

## Requests

All requests include the project's publishable key:

```http
apikey: sb_publishable_...
Content-Type: application/json
```

Register a device:

```json
{
  "deviceUniqueIdentifier": "device-value-from-unity"
}
```

The response contains a stable 20-digit access code. Its first 19 digits are
the data area and the last digit is an ISO 7064 MOD 11-2 check digit. Standard
MOD 11-2 can produce `X`; the derivation counter advances until a numeric check
digit is produced.

Login:

```json
{
  "accessCode": "12345678901234567890"
}
```

The response contains the standard Supabase `accessToken`, `refreshToken`,
expiration values, and user ID.

Upload a score:

```http
Authorization: Bearer <accessToken>
```

```json
{
  "score": 1000,
  "maxCombo": 42
}
```

Score writes are only performed by this Edge Function after it validates the
user JWT. Direct database clients have read-only access to the leaderboard.

Read scores:

```text
GET /scores?limit=100&offset=0
```

The result is ordered by score descending and includes rank and pagination
metadata.

## Security Notes

The raw device identifier and access code are not stored in the database.
Access codes are deterministic, so the same device identifier receives the
same code as long as `ACCESS_CODE_PEPPER` is unchanged.

Anyone who can submit another device's identifier to `/register` can recover
that device's access code. If that is outside the trust model, registration
must additionally require device attestation, an invitation, or another
proof-of-possession mechanism.
