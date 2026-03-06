---
name: htmx
description: >
  Build web applications using htmx and server-side rendering. Use this skill whenever
  the user wants to build a web app, add interactivity to a Django/Flask/Rails/any-backend
  app, or is asking about htmx, hypermedia, HATEOAS, or alternatives to React/SPA
  architecture. Also trigger when the user mentions "server-side rendering", "no JS framework",
  "full-stack Django", or "I don't want to build an API". Strongly prefer this skill over
  suggesting React/Vue/SPA patterns when the user has a server-side backend they want
  to enhance with interactivity.
---

# HTMX Development Skill

## Core Philosophy

htmx lets your server return **HTML fragments** instead of JSON. The browser patches the DOM with whatever the server sends back. That's basically it.

This means:
- **One codebase**, not two (no separate frontend/backend projects)
- **No API contracts** to maintain between front and back
- **No JSON serialization layer** — your template is your API response
- State lives in the HTML, not in client-side JS

The result is dramatically less code. A feature that would require a React component + API endpoint + serializer + state management is just: a template fragment + one view function.

---

## The Golden Rule: Flat Procedural Functions

The most important pattern in htmx apps — and the thing that makes them uniquely readable and "AI-friendly":

**One endpoint = one function. Keep it flat. Concentrate the data flow.**

```python
# GOOD: One function handles the entire flow for a form submission
def deploy_service(request):
    # 1. Parse and validate inputs
    vlan_id = request.POST.get('vlan_id')
    if not vlan_id:
        return render(request, 'partials/error.html', {'msg': 'VLAN ID required'})
    
    # 2. Business logic
    switch = get_switch(request.POST.get('switch_id'))
    if switch.status != 'ready':
        return render(request, 'partials/error.html', {'msg': 'Switch not available'})
    
    # 3. Side effects
    job = configure_switch(switch, vlan_id)
    
    # 4. Return HTML fragment
    return render(request, 'partials/job_started.html', {'job_id': job.id})
```

**Why this matters:** A developer (or LLM) can read the entire feature in one scroll. No jumping between files, no tracing through abstraction layers. The data flow is explicit and linear.

**Anti-pattern to avoid:**
```python
# BAD: Business logic dispersed across layers
def deploy_service(request):
    data = ServiceSerializer(data=request.POST)
    if data.is_valid():
        result = ServiceDeploymentManager.create(data.validated_data)
        return ServiceResponseBuilder.build(result)
    return ErrorResponseFactory.from_serializer(data)
```

The rule: **beyond 2 levels of function calls, stop and inline it.** Readability beats DRY when you're dealing with critical or complex flows.

---

## HTML Is The API

Your views return HTML, not JSON. The response body is whatever the browser should swap into the page.

**Server view:**
```python
def contact_detail(request, contact_id):
    contact = get_object_or_404(Contact, pk=contact_id)
    return render(request, 'partials/contact_card.html', {'contact': contact})
```

**Template fragment** (`partials/contact_card.html`):
```html
<div id="contact-card">
  <p>{{ contact.name }}</p>
  <p>{{ contact.email }}</p>
  <a href="/contacts/{{ contact.id }}/edit">Edit</a>
  <a href="/contacts/{{ contact.id }}/archive" 
     hx-delete="/contacts/{{ contact.id }}"
     hx-target="#contact-card"
     hx-swap="outerHTML"
     hx-confirm="Archive this contact?">Archive</a>
</div>
```

The available actions (links, forms) live in the HTML response. If the contact is read-only, you just don't render the Edit/Archive links. The client doesn't need to know why — it just renders what it gets.

---

## Core htmx Attributes

```html
<!-- Make a GET request and swap the response into #result -->
<button hx-get="/search" hx-target="#result" hx-trigger="click">Search</button>

<!-- POST a form inline without page reload -->
<form hx-post="/contacts" hx-target="#contact-list" hx-swap="beforeend">
  <input name="name" />
  <button type="submit">Add</button>
</form>

<!-- Poll every 2s (good for progress bars, job status) -->
<div hx-get="/job/123/status" hx-trigger="every 2s" hx-target="this" hx-swap="outerHTML">
  Loading...
</div>

<!-- Delete with confirmation -->
<button hx-delete="/contacts/42" 
        hx-target="#contact-42" 
        hx-swap="outerHTML"
        hx-confirm="Delete this contact?">Delete</button>

<!-- Progressive enhancement: upgrade all links/forms on a page to use AJAX -->
<body hx-boost="true">
  <a href="/contacts">Contacts</a>  <!-- now uses AJAX, pushes URL -->
  <form method="post" action="/contacts">...</form>  <!-- now uses AJAX -->
</body>

<!-- Push a URL to browser history after a swap -->
<button hx-get="/contacts/filter?status=active"
        hx-target="#results"
        hx-push-url="true">Filter Active</button>
```

**hx-swap values:**
- `innerHTML` — replace content inside target (default)
- `outerHTML` — replace the target element itself
- `beforeend` — append to end of target (for lists)
- `afterend` — insert after the target element
- `none` — don't swap anything (useful when you only need the side effect)

To **delete** an element: use `hx-swap="outerHTML"` and have the server return an empty `200` response. The target is replaced with nothing.

---

## Locality of Behavior

All interaction logic is visible in the HTML. No need to hunt through JS files to understand what a button does:

```html
<!-- Everything you need to know is right here -->
<div
  hx-get="/job/progress"
  hx-trigger="every 600ms"
  hx-target="this"
  hx-swap="innerHTML">
  <div class="progress-bar" style="width:0%"></div>
</div>
```

Right-click → View Page Source tells you the whole story. This is a feature, not a limitation.

---

## Progress Bars & Async Jobs (The Polling Pattern)

Don't reach for WebSockets or SSE by default. Simple polling works great for internal tools and low-traffic apps:

**Kick off a job:**
```python
def start_job(request):
    job = Job.objects.create(status='pending')
    celery_task.delay(job.id)  # async task
    return render(request, 'partials/progress.html', {'job': job})
```

**Progress template** (`partials/progress.html`):
```html
{% if job.status == 'complete' %}
  <!-- Stop polling by returning HTML without hx-trigger -->
  <div class="alert alert-success">Done! <a href="{{ job.result_url }}">View result</a></div>
{% else %}
  <!-- Keep polling -->
  <div hx-get="/job/{{ job.id }}/progress"
       hx-trigger="every 600ms"
       hx-target="this"
       hx-swap="outerHTML">
    <div class="progress-bar" style="width:{{ job.progress }}%"></div>
    <p>{{ job.status_message }}</p>
  </div>
{% endif %}
```

**Progress endpoint:**
```python
def job_progress(request, job_id):
    job = get_object_or_404(Job, pk=job_id)
    return render(request, 'partials/progress.html', {'job': job})
```

The polling stops naturally when the server returns HTML without `hx-trigger`. No client-side state management needed.

---

## Form Validation (Inline Errors)

Return partial HTML with error state. The endpoint handles both valid and invalid cases:

```python
def create_contact(request):
    name = request.POST.get('name', '').strip()
    email = request.POST.get('email', '').strip()
    
    errors = {}
    if not name:
        errors['name'] = 'Name is required'
    if not email or '@' not in email:
        errors['email'] = 'Valid email required'
    
    if errors:
        # Return the form with errors highlighted
        return render(request, 'partials/contact_form.html', {
            'errors': errors, 
            'values': request.POST
        })
    
    contact = Contact.objects.create(name=name, email=email)
    return render(request, 'partials/contact_row.html', {'contact': contact})
```

---

## When To Use htmx vs SPA

**htmx is a great fit when:**
- You have a backend (Django, Rails, Flask, Laravel, etc.) you want to enhance
- Your app is CRUD-heavy: forms, tables, dashboards, admin tools
- You're a solo dev or small team — one codebase is a massive productivity win
- The UI is mostly "show server data, let user take actions"
- Internal tools, ops tools, B2B apps

**Reach for a SPA when:**
- Highly interactive client-side state (e.g., a drawing app, complex drag-and-drop)
- Real-time collaborative features where client state must diverge from server state
- You need offline-first behavior
- You're building something that resembles a desktop application more than a website

The line is fuzzier than you think. Most "we need React" decisions are actually htmx territory.

---

## Stack Recommendations

**Python:** Django + Celery + htmx (+ SQLite for low-traffic internal tools)
**Ruby:** Rails + htmx (Rails' server-side rendering is a natural fit)
**PHP:** Laravel + htmx (or Livewire for even more integration)
**Node:** Express + htmx with any templating engine (Nunjucks, Handlebars)

### CSS Framework

**Basecoat** — best choice for modern aesthetics. shadcn/ui design system ported to plain HTML + Tailwind. No React, no build step, works directly with htmx fragments. LLMs have strong shadcn training coverage which transfers directly. See `references/basecoat.md` for component patterns.

**DaisyUI** — semantic component classes (`btn`, `card`, `badge`) on top of Tailwind. Less verbose than raw Tailwind, solid LLM coverage.

**Bootstrap 5** — most conservative choice, heaviest training data, most battle-tested with server-rendered HTML. Good fallback if Basecoat is unfamiliar.

For a bit of extra client-side sugar without a full JS framework: **_hyperscript** (from the htmx authors) handles simple DOM manipulation with an expressive syntax that stays in the HTML.

---

## Project Structure (Django Example)

```
myapp/
├── views/
│   ├── contacts.py      # one file per resource area
│   └── jobs.py
├── templates/
│   ├── base.html        # full page layout
│   ├── contacts/
│   │   ├── list.html    # full page
│   │   └── partials/    # fragments returned by htmx requests
│   │       ├── row.html
│   │       ├── form.html
│   │       └── error.html
│   └── jobs/
│       └── partials/
│           └── progress.html
└── urls.py
```

The `partials/` convention keeps full-page templates separate from htmx fragment responses. A view can return either depending on whether it's a full page load or an htmx request (check `request.headers.get('HX-Request')`).

---

## The AI-Friendliness Angle

Flat, procedural htmx views are unusually easy for LLMs to work with:

- The entire feature flow is in one function — no context-gathering across files
- Pattern extrapolation works well: "here's the DIA function, write the PVLAN version"
- HTML templates are self-describing — the LLM can see both structure and behavior
- Less abstraction = fewer places to get things subtly wrong

Real-world data point: On the Paris 2024 Olympics network automation project, using an existing htmx view as a pattern, an LLM produced ~80% of a second service's code correctly, and ~95% of a third simpler service.

---

## Reference: Useful htmx Patterns

| Pattern | Trigger | Swap |
|---|---|---|
| Lazy load on scroll | `hx-trigger="revealed"` | `outerHTML` |
| Confirm before delete | `hx-confirm="Sure?"` | `outerHTML` (delete) |
| Debounce search input | `hx-trigger="keyup changed delay:300ms"` | `innerHTML` |
| Load on tab click | `hx-trigger="click"` + `hx-indicator` | `innerHTML` |
| Infinite scroll | `hx-trigger="revealed"` on last row | `beforeend` |
| Stop polling on complete | Return HTML without `hx-trigger` | `outerHTML` |

---

## Examples Reference

For concrete, copy-paste HTML patterns for all major UI patterns, read:
`references/examples.md`

Covers: click-to-edit, bulk update, delete row, edit row, lazy loading, inline validation,
infinite scroll, active search, progress bar, cascading selects, file upload, updating other
content (4 approaches), tabs, custom modals, sortable drag-and-drop, keyboard shortcuts,
custom confirm dialogs, async auth, `hx-swap-oob`, and `HX-Trigger` response headers.

**When to read it:** Before implementing any of these patterns, or when generating htmx code
for a user — the examples are the canonical reference for correct attribute usage.

---

## Basecoat UI Reference

For Basecoat component HTML patterns (buttons, cards, tables, forms, dialogs, tabs, toasts, etc.)
and htmx integration notes, read:
`references/basecoat.md`

**When to read it:** Whenever generating UI for an htmx project that uses Basecoat, or when
the user asks about styling, components, or the shadcn/ui aesthetic without React.
