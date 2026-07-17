# Simple Job Tracker API

Base URL: `http://localhost:3000`

## Authentication

All endpoints except `/api/auth/register` and `/api/auth/login` require a Bearer token.

```
Authorization: Bearer <token>
```

Tokens expire after 72 hours.

---

## Auth Endpoints

### POST /api/auth/register

Create a new user account.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "securepassword",
  "name": "John Doe"
}
```

**Response `201`:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "name": "John Doe",
  "created_at": "2026-07-02T10:00:00Z",
  "updated_at": "2026-07-02T10:00:00Z"
}
```

**Errors:** `400` (missing fields), `409` (email exists)

---

### POST /api/auth/login

Authenticate and receive a JWT.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "securepassword"
}
```

**Response `200`:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "name": "John Doe",
    "created_at": "2026-07-02T10:00:00Z",
    "updated_at": "2026-07-02T10:00:00Z"
  }
}
```

**Errors:** `401` (invalid email or password)

---

### GET /api/auth/me

Return the authenticated user's profile.

**Headers:** `Authorization: Bearer <token>`

**Response `200`:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "name": "John Doe",
  "created_at": "2026-07-02T10:00:00Z",
  "updated_at": "2026-07-02T10:00:00Z"
}
```

**Errors:** `401` (invalid/missing token), `404` (user not found)

---

## Job Application Endpoints

All require `Authorization: Bearer <token>`.

### GET /api/applications

List all job applications for the authenticated user.

Ordered by `apply_date_time` descending (nulls last), then `created_at` descending.

**Response `200`:**
```json
[
  {
    "id": "660e8400-e29b-41d4-a716-446655440001",
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "Senior Frontend Engineer",
    "company": "Acme Corp",
    "category": "Engineering",
    "description": "React, TypeScript, etc.",
    "tech_stack": ["React", "TypeScript", "GraphQL"],
    "status": "applied",
    "apply_date_time": "2026-07-01T08:00:00Z",
    "created_at": "2026-07-01T10:00:00Z",
    "updated_at": "2026-07-01T10:00:00Z"
  }
]
```

---

### GET /api/applications/:id

Get a single job application.

**Response `200`:** Same shape as a single item above.

**Errors:** `404` (not found or not owned by user)

---

### POST /api/applications

Create a new job application.

**Request:**
```json
{
  "title": "Senior Frontend Engineer",
  "company": "Acme Corp",
  "category": "Engineering",
  "description": "React, TypeScript position",
  "tech_stack": ["React", "TypeScript", "GraphQL"],
  "status": "applied",
  "apply_date_time": "2026-07-01T08:00:00Z"
}
```

| Field            | Type     | Required | Default     |
|------------------|----------|----------|-------------|
| `title`          | string   | yes      |             |
| `company`        | string   | yes      |             |
| `category`       | string   | no       | `null`      |
| `description`    | string   | no       | `null`      |
| `tech_stack`     | string[] | no       | `null`      |
| `status`         | string   | no       | `"applied"` |
| `apply_date_time`| string   | no       | `null`      |

`apply_date_time` must be in RFC3339 format (e.g. `2026-07-01T08:00:00Z`).

**Response `201`:** Full application object with generated id and timestamps.

**Errors:** `400` (missing title/company, invalid date format), `401` (unauthorized)

---

### PUT /api/applications/:id

Update an existing job application (partial update — only send the fields to change).

**Request:**
```json
{
  "status": "interview",
  "tech_stack": ["React", "TypeScript", "GraphQL", "Node.js"]
}
```

All fields are optional:

| Field            | Type     | Notes                                       |
|------------------|----------|---------------------------------------------|
| `title`          | string   |                                             |
| `company`        | string   |                                             |
| `category`       | string   | Pass `null` to clear                        |
| `description`    | string   | Pass `null` to clear                        |
| `tech_stack`     | string[] | Replaces entire array if provided           |
| `status`         | string   |                                             |
| `apply_date_time`| string   | RFC3339 format. Pass `null` to clear.       |

**Response `200`:** Full updated application object.

**Errors:** `404` (not found or not owned by user)

---

### DELETE /api/applications/:id

Delete a job application.

**Response `204`:** No content (success).

**Errors:** `404` (not found or not owned by user)

---

## Job Application Progress History (Events)

Each job application has a history of progress events (e.g. Applied on July 12th, Interview on July 17th). An event is automatically recorded when an application is created, and whenever its `status` changes via `PUT /api/applications/:id` or via the events endpoint below.

All require `Authorization: Bearer <token>`.

### GET /api/applications/:id/events

List the progress history for a job application, ordered by `event_date` ascending.

**Response `200`:**
```json
[
  {
    "id": "770e8400-e29b-41d4-a716-446655440002",
    "job_application_id": "660e8400-e29b-41d4-a716-446655440001",
    "status": "applied",
    "note": null,
    "event_date": "2026-07-12T08:00:00Z",
    "created_at": "2026-07-12T08:00:00Z"
  },
  {
    "id": "880e8400-e29b-41d4-a716-446655440003",
    "job_application_id": "660e8400-e29b-41d4-a716-446655440001",
    "status": "interview",
    "note": "Phone screen with recruiter",
    "event_date": "2026-07-17T15:00:00Z",
    "created_at": "2026-07-17T15:05:00Z"
  }
]
```

**Errors:** `404` (application not found or not owned by user)

---

### POST /api/applications/:id/events

Add a progress event to a job application's history. This also updates the parent application's `status` field to match.

**Request:**
```json
{
  "status": "interview",
  "note": "Onsite interview scheduled",
  "event_date": "2026-07-17T15:00:00Z"
}
```

| Field        | Type   | Required | Default          |
|--------------|--------|----------|------------------|
| `status`     | string | yes      |                  |
| `note`       | string | no       | `null`           |
| `event_date` | string | no       | current time     |

`event_date` must be in RFC3339 format (e.g. `2026-07-17T15:00:00Z`).

**Response `201`:** The created event object.

**Errors:** `400` (missing status, invalid date format), `404` (application not found or not owned by user)

---

### DELETE /api/applications/:id/events/:eventId

Remove an event from a job application's history.

**Response `204`:** No content (success).

**Errors:** `404` (application or event not found)

---

## Status Values

Suggested statuses (free-text field, no enum enforcement):

| Value          | Meaning                  |
|----------------|--------------------------|
| `applied`      | Application submitted    |
| `screening`    | Initial screening        |
| `interview`    | Interview stage          |
| `offer`        | Offer received           |
| `rejected`     | Not moving forward       |
| `accepted`     | Offer accepted           |
| `withdrawn`    | Withdrew application     |

---

## Error Response Format

```json
{
  "error": "description of what went wrong"
}
```
