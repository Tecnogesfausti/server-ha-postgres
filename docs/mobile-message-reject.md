# Mobile message rejection

`POST /api/mobile/v1/message/{id}/reject` (also `/messages/{id}/reject`) requires `Authorization: Bearer <device token>`. It only accesses an order assigned to that device. The JSON body contains one required `reason`:

- `wrong_recipient`
- `unsupported_format`
- `unsupported_encryption`
- `unsupported_sim`
- `unsupported_command`

A successful commit returns HTTP 200 and JSON `{"id":"<id>","state":"Failed","reason":"wrong_recipient"}`. The reason is stored on the message and as the recipient error, and a `Failed` state-history record is retained. A pending order is no longer returned by the mobile pending list. Concurrent or later repeats with the same ID and reason return the same success without adding history. A repeated request after a lost response is safe.

Responses: 400 for invalid ID, JSON or reason; 401 for absent or invalid device token; 404 for a missing order or an order assigned to a different device; 409 if the order is no longer pending, or was rejected for another reason; 500 for a persistence failure. A 200 is sent only after the transaction has committed.

This route does not alter the existing bulk PATCH endpoint. Clients must not send further PATCH updates for a rejected order: the existing PATCH mechanism can change its state independently. The change adds a nullable `rejection_reason` column using the server's database migration at startup. Back up and test migration and rollback before upgrading an existing instance. SMS bodies, tokens and phone numbers should not be logged.
