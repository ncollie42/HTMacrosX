# Basecoat UI Reference

Basecoat is shadcn/ui ported to plain HTML + Tailwind CSS. No React, no build step, works with
any server-side stack (Django, Rails, Flask, Laravel). Full docs: https://basecoatui.com

---

## Installation (CDN — fastest for htmx projects)

```html
<head>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/basecoat-css@0.3.11/dist/basecoat.cdn.min.css">
  <!-- All interactive components (only ~3kB gzipped) -->
  <script src="https://cdn.jsdelivr.net/npm/basecoat-css@0.3.11/dist/js/all.min.js" defer></script>
</head>
```

Or cherry-pick JS for specific interactive components (Dropdown Menu, Popover, Select, Sidebar, Tabs, Toast):
```html
<script src="https://cdn.jsdelivr.net/npm/basecoat-css@0.3.11/dist/js/basecoat.min.js" defer></script>
<script src="https://cdn.jsdelivr.net/npm/basecoat-css@0.3.11/dist/js/tabs.min.js" defer></script>
```

## Tailwind project setup

```css
/* In your CSS file */
@import "basecoat-css";
```

Fully compatible with shadcn/ui themes — import any theme from ui.shadcn.com/themes.

---

## Components That Need JS vs Pure CSS

**Pure CSS (no JS needed):** Button, Badge, Card, Alert, Table, Input, Field, Form,
Checkbox, Radio Group, Progress, Skeleton, Spinner, Breadcrumb, Pagination, Avatar, Kbd, Label, Textarea, Switch, Slider

**Requires JS:** Dropdown Menu, Popover, Select, Sidebar, Tabs, Toast

---

## Buttons

```html
<!-- Variants -->
<button class="btn">Default</button>
<button class="btn-primary">Primary</button>
<button class="btn-outline">Outline</button>
<button class="btn-destructive">Destructive</button>
<button class="btn-ghost">Ghost</button>
<button class="btn-link">Link</button>

<!-- Sizes -->
<button class="btn-sm">Small</button>
<button class="btn">Default</button>
<button class="btn-lg">Large</button>

<!-- Icon-only buttons -->
<button class="btn-icon">...</button>
<button class="btn-icon-outline">...</button>
<button class="btn-sm-icon-destructive">...</button>

<!-- With icon + text (use Lucide SVGs) -->
<button class="btn">
  <svg ...>...</svg>
  Send email
</button>

<!-- Loading state -->
<button class="btn-outline" disabled>
  <svg class="animate-spin" ...>...</svg>
  Loading
</button>

<!-- As link -->
<a href="/contacts" class="btn-primary">Go to Contacts</a>
```

---

## Badge

```html
<span class="badge">Default</span>
<span class="badge-primary">Primary</span>
<span class="badge-secondary">Secondary</span>
<span class="badge-destructive">Destructive</span>
<span class="badge-outline">Outline</span>

<!-- As link -->
<a href="#" class="badge-outline">Link →</a>

<!-- With icon -->
<span class="badge-destructive">
  <svg ...>...</svg>
  Error
</span>
```

---

## Card

Semantic structure: `<div class="card">` → `<header>` → `<section>` → `<footer>`

```html
<div class="card">
  <header>
    <h2>Card Title</h2>
    <p>Optional description text.</p>
  </header>
  <section>
    <!-- Main content -->
    <p>Card body goes here.</p>
  </section>
  <footer class="flex items-center justify-between">
    <!-- Footer actions -->
    <button class="btn-outline">Cancel</button>
    <button class="btn-primary">Save</button>
  </footer>
</div>
```

---

## Table

```html
<div class="overflow-x-auto">
  <table class="table">
    <caption>A list of recent invoices.</caption>
    <thead>
      <tr>
        <th>Invoice</th>
        <th>Status</th>
        <th>Method</th>
        <th class="text-right">Amount</th>
      </tr>
    </thead>
    <tbody>
      <tr>
        <td class="font-medium">INV001</td>
        <td><span class="badge-outline">Paid</span></td>
        <td>Credit Card</td>
        <td class="text-right">$250.00</td>
      </tr>
    </tbody>
  </table>
</div>
```

---

## Form Fields

The `field` class wires together label, input, and description/error automatically.

```html
<form class="form space-y-6">
  <!-- Basic field -->
  <div role="group" class="field">
    <label for="email">Email</label>
    <input id="email" type="email" placeholder="you@example.com">
  </div>

  <!-- With description -->
  <div role="group" class="field">
    <label for="username">Username</label>
    <input id="username" type="text" aria-describedby="username-desc">
    <p id="username-desc">Must be at least 3 characters.</p>
  </div>

  <!-- With validation error -->
  <div role="group" class="field">
    <label for="slug">Slug</label>
    <input id="slug" type="text" aria-invalid="true" aria-describedby="slug-error">
    <p id="slug-error" role="alert">That slug is already taken.</p>
  </div>

  <button type="submit" class="btn">Submit</button>
</form>
```

**Fieldset grouping:**
```html
<fieldset class="fieldset">
  <legend>Profile</legend>
  <p>This information will be displayed publicly.</p>
  <div role="group" class="field">
    <label for="name">Full name</label>
    <input id="name" type="text">
  </div>
</fieldset>
```

---

## Input, Textarea, Select

Styled automatically inside `.form` or with explicit class:

```html
<!-- Input -->
<input type="text" class="input" placeholder="Search...">

<!-- Textarea -->
<textarea class="input" rows="4"></textarea>

<!-- Native select (pure CSS) -->
<select class="select">
  <option value="">Choose...</option>
  <option value="a">Option A</option>
</select>

<!-- Checkbox -->
<label class="label">
  <input type="checkbox" class="input"> Remember me
</label>

<!-- Radio group -->
<fieldset class="grid gap-3">
  <label class="label"><input type="radio" name="plan" value="free" class="input"> Free</label>
  <label class="label"><input type="radio" name="plan" value="pro" class="input"> Pro</label>
</fieldset>
```

---

## Alert / Notification

```html
<div role="alert" class="alert">
  <svg ...>...</svg>  <!-- icon -->
  <div>
    <h3>Heads up!</h3>
    <p>You can add components to your app using the CLI.</p>
  </div>
</div>

<!-- Variants -->
<div role="alert" class="alert-destructive">...</div>
```

---

## Dialog / Modal

Uses native `<dialog>` element. No custom JS needed for basic open/close.

```html
<!-- Trigger -->
<button type="button" class="btn-outline"
        onclick="document.getElementById('my-dialog').showModal()">
  Open Dialog
</button>

<!-- Dialog -->
<dialog id="my-dialog" class="dialog"
        aria-labelledby="dialog-title"
        aria-describedby="dialog-desc"
        onclick="if(event.target===this)this.close()">
  <article onclick="event.stopPropagation()">
    <header>
      <h2 id="dialog-title">Edit Profile</h2>
      <p id="dialog-desc">Make changes to your profile here.</p>
      <button type="button" class="btn-icon-ghost"
              onclick="document.getElementById('my-dialog').close()">
        <svg ...><!-- X icon --></svg>
      </button>
    </header>
    <section>
      <!-- Form content -->
    </section>
    <footer>
      <button class="btn-outline"
              onclick="document.getElementById('my-dialog').close()">Cancel</button>
      <button class="btn-primary">Save changes</button>
    </footer>
  </article>
</dialog>
```

**Alert Dialog** (no backdrop close, no X button):
```html
<dialog id="confirm-dialog" class="dialog" aria-labelledby="confirm-title">
  <div>
    <header>
      <h2 id="confirm-title">Are you absolutely sure?</h2>
      <p>This action cannot be undone.</p>
    </header>
    <footer>
      <button class="btn-outline"
              onclick="document.getElementById('confirm-dialog').close()">Cancel</button>
      <button class="btn-primary"
              onclick="document.getElementById('confirm-dialog').close()">Continue</button>
    </footer>
  </div>
</dialog>
```

---

## Tabs (requires JS)

```html
<div class="tabs" id="my-tabs">
  <nav role="tablist" aria-orientation="horizontal">
    <button type="button" role="tab"
            id="my-tabs-tab-1"
            aria-controls="my-tabs-panel-1"
            aria-selected="true"
            tabindex="0">Account</button>
    <button type="button" role="tab"
            id="my-tabs-tab-2"
            aria-controls="my-tabs-panel-2"
            aria-selected="false"
            tabindex="-1">Password</button>
  </nav>

  <div role="tabpanel"
       id="my-tabs-panel-1"
       aria-labelledby="my-tabs-tab-1"
       tabindex="-1">
    <!-- Tab 1 content -->
    <div class="card">...</div>
  </div>

  <div role="tabpanel"
       id="my-tabs-panel-2"
       aria-labelledby="my-tabs-tab-2"
       tabindex="-1"
       hidden>
    <!-- Tab 2 content -->
    <div class="card">...</div>
  </div>
</div>
```

**With htmx (HATEOAS tabs):** Skip Basecoat's JS tabs entirely. Use htmx to load tab content
from the server. Each tab response includes the full tab bar with the active tab marked —
see the htmx tabs pattern in `references/examples.md`.

---

## Toast (requires JS)

Add toaster to end of `<body>`:
```html
<body>
  <!-- page content -->
  <div data-toaster></div>
</body>
```

Trigger from server via `HX-Trigger` response header (best for htmx):
```python
# Django: fire a toast after a successful form submission
response = render(request, 'partials/contact_row.html', {'contact': contact})
response['HX-Trigger'] = json.dumps({
    "basecoat:toast": {
        "config": {
            "category": "success",
            "title": "Contact saved",
            "description": "The contact was added successfully."
        }
    }
})
return response
```

Or trigger from front-end JS:
```html
<button onclick="document.dispatchEvent(new CustomEvent('basecoat:toast', {
  detail: { config: { category: 'success', title: 'Done!', description: 'Changes saved.' } }
}))">
  Save
</button>
```

Toast categories: `success`, `error`, `warning`, `info` (default)

---

## Accordion (pure CSS with `<details>`)

```html
<section class="accordion">
  <details class="group border-b last:border-b-0">
    <summary class="w-full focus-visible:ring-ring/50 transition-all outline-none rounded-md">
      <h2 class="flex flex-1 items-start justify-between gap-4 py-4 text-sm font-medium hover:underline">
        Is it accessible?
        <svg class="text-muted-foreground size-4 shrink-0 transition-transform duration-200 group-open:rotate-180"
             ...><!-- chevron down --></svg>
      </h2>
    </summary>
    <section class="pb-4">
      <p class="text-sm">Yes. It adheres to the WAI-ARIA design pattern.</p>
    </section>
  </details>
  <!-- More <details> items... -->
</section>
```

---

## Progress / Spinner / Skeleton

```html
<!-- Progress bar -->
<div class="progress" role="progressbar" aria-valuenow="60" aria-valuemin="0" aria-valuemax="100">
  <div style="width: 60%"></div>
</div>

<!-- Spinner (inline loading indicator) -->
<span class="spinner" aria-label="Loading..."></span>

<!-- Skeleton (loading placeholder) -->
<div class="space-y-3">
  <div class="skeleton h-4 w-3/4"></div>
  <div class="skeleton h-4 w-1/2"></div>
  <div class="skeleton h-4 w-full"></div>
</div>
```

---

## Sidebar (requires JS)

```html
<aside class="sidebar" id="main-sidebar" data-side="left">
  <nav>
    <header>
      <a href="/" class="flex items-center gap-2 font-semibold">
        My App
      </a>
    </header>
    <ul>
      <li><a href="/dashboard">Dashboard</a></li>
      <li><a href="/contacts" aria-current="page">Contacts</a></li>
      <li><a href="/settings">Settings</a></li>
    </ul>
  </nav>
</aside>

<!-- Toggle button (can be anywhere on page) -->
<button type="button"
        onclick="document.dispatchEvent(new CustomEvent('basecoat:sidebar', { detail: { id: 'main-sidebar' } }))">
  Toggle Sidebar
</button>
```

---

## Pagination

```html
<nav aria-label="Pagination" class="pagination">
  <a href="?page=1" class="btn-ghost btn-icon" aria-label="First page">«</a>
  <a href="?page=2" class="btn-ghost btn-icon" aria-label="Previous page">‹</a>

  <a href="?page=1" class="btn-ghost">1</a>
  <a href="?page=2" class="btn-ghost" aria-current="page">2</a>
  <a href="?page=3" class="btn-ghost">3</a>

  <a href="?page=3" class="btn-ghost btn-icon" aria-label="Next page">›</a>
  <a href="?page=10" class="btn-ghost btn-icon" aria-label="Last page">»</a>
</nav>
```

---

## Avatar

Avatars are plain `<img>` elements with Tailwind utility classes — no dedicated CSS class.

```html
<!-- Single avatar (circle) -->
<img class="size-8 shrink-0 object-cover rounded-full" alt="@username" src="/avatar.png" />

<!-- Rounded square -->
<img class="size-8 shrink-0 object-cover rounded-lg" alt="@username" src="/avatar.png" />

<!-- Stacked group (overlapping avatars) -->
<div class="flex -space-x-2 [&_img]:ring-background [&_img]:ring-2 [&_img]:size-8 [&_img]:shrink-0 [&_img]:object-cover [&_img]:rounded-full">
  <img alt="@user1" src="/avatar1.png" />
  <img alt="@user2" src="/avatar2.png" />
  <img alt="@user3" src="/avatar3.png" />
</div>

<!-- Fallback initials (no image) -->
<div class="size-8 rounded-full bg-muted flex items-center justify-center text-sm font-medium">CN</div>
```

---

## Accordion

No dedicated `accordion` class — uses native `<details>`/`<summary>` HTML with Tailwind:

```html
<section class="accordion">
  <details class="group border-b last:border-b-0">
    <summary class="w-full focus-visible:border-ring py-4 cursor-pointer list-none">
      <h2 class="flex flex-1 items-start justify-between text-sm font-medium">
        Is it accessible?
        <!-- chevron icon rotates via group-open -->
        <svg class="shrink-0 transition-transform group-open:rotate-180" ...>...</svg>
      </h2>
    </summary>
    <section class="pb-4">
      <p class="text-sm">Yes. It adheres to the WAI-ARIA design pattern.</p>
    </section>
  </details>
  <details class="group border-b last:border-b-0">
    <!-- repeat pattern -->
  </details>
</section>
```

The `group-open:rotate-180` on the chevron handles the open/close animation via CSS only.

---

## Tooltip

No dedicated class — uses `data-tooltip` attribute anywhere:

```html
<!-- Basic -->
<button class="btn" data-tooltip="Save changes">Save</button>

<!-- With side and alignment -->
<button class="btn-outline" data-tooltip="Tooltip text" data-side="top" data-align="center">Top</button>
<button class="btn-outline" data-tooltip="Tooltip text" data-side="bottom" data-align="start">Bottom + Start</button>
<button class="btn-outline" data-tooltip="Tooltip text" data-side="left">Left</button>
<button class="btn-outline" data-tooltip="Tooltip text" data-side="right">Right</button>
```

`data-side`: `top` | `bottom` | `left` | `right` (default: top)
`data-align`: `start` | `center` | `end` (default: center)

Tooltip JS is part of `all.min.js` (no separate script needed).

---

## Switch

Switch is `<input type="checkbox" role="switch">` — no dedicated CSS class needed. Style inside `.field` or use `.form` parent.

```html
<!-- Standalone switch -->
<label class="flex items-center gap-2">
  <input type="checkbox" role="switch" id="airplane-mode">
  <span>Airplane Mode</span>
</label>

<!-- Disabled -->
<label class="flex items-center gap-2 opacity-50">
  <input type="checkbox" role="switch" id="bluetooth" checked disabled>
  <span>Bluetooth</span>
</label>

<!-- In a field (horizontal layout) -->
<div role="group" class="field" data-orientation="horizontal">
  <section>
    <label for="notifications">Enable notifications</label>
    <p>You can disable this at any time.</p>
  </section>
  <input id="notifications" type="checkbox" role="switch">
</div>
```

---

## Combobox

Combobox uses the exact same markup as `Select`, but adds a search `<header>` inside the popover. The JS is the same; Basecoat detects the search input to enable filtering.

```html
<div id="my-combobox" class="select">
  <button type="button" class="btn-outline w-[200px]"
    id="my-combobox-trigger"
    aria-haspopup="listbox"
    aria-expanded="false"
    aria-controls="my-combobox-listbox">
    <span class="truncate"></span>
    <!-- chevrons-up-down icon -->
  </button>
  <div id="my-combobox-popover" data-popover aria-hidden="true">
    <!-- This header with search input is what makes it a combobox -->
    <header>
      <!-- search icon -->
      <input type="text" placeholder="Search..." />
    </header>
    <div role="listbox" id="my-combobox-listbox" aria-labelledby="my-combobox-trigger">
      <div role="option" data-value="nextjs">Next.js</div>
      <div role="option" data-value="svelte">SvelteKit</div>
      <hr role="separator">
      <!-- footer action -->
      <div role="menuitem">+ Create new</div>
    </div>
  </div>
</div>
```

Grouped options: wrap in `<div role="group" aria-labelledby="group-heading-id">` with a `<div role="heading">` inside.

---

## Dropdown Menu

Uses the `.dropdown-menu` class wrapper. Trigger is a `<button>` with `aria-haspopup="menu"`. Content uses `role="menu"` with `role="menuitem"` children.

```html
<div id="my-dropdown" class="dropdown-menu">
  <button type="button" class="btn-outline"
    id="my-dropdown-trigger"
    aria-haspopup="menu"
    aria-expanded="false"
    aria-controls="my-dropdown-menu">
    Open
  </button>
  <div id="my-dropdown-popover" data-popover aria-hidden="true" class="min-w-56">
    <div role="menu" id="my-dropdown-menu" aria-labelledby="my-dropdown-trigger">
      <!-- Optional group with heading -->
      <div role="group" aria-labelledby="account-heading">
        <div role="heading" id="account-heading">My Account</div>
        <div role="menuitem">
          Profile
          <span class="text-muted-foreground ml-auto text-xs tracking-widest">⇧⌘P</span>
        </div>
        <div role="menuitem">Settings</div>
        <div role="menuitem" disabled>Disabled item</div>
      </div>
      <hr role="separator">
      <div role="menuitem">Logout</div>
    </div>
  </div>
</div>
```

For checkboxes in dropdown: add `<input type="checkbox">` inside `role="menuitem"`.
For radio group: wrap items in `<div role="radiogroup">` with `role="menuitemradio"`.
Trigger can be an avatar button — just use `class="btn-icon-ghost rounded-full"` on the button.

---

## Button Group

No dedicated Basecoat class — use Tailwind flexbox. Group buttons by removing border-radius on interior edges:

```html
<!-- Simple button group -->
<div class="flex" role="group" aria-label="Text formatting">
  <button class="btn-outline rounded-r-none border-r-0">Bold</button>
  <button class="btn-outline rounded-none border-r-0">Italic</button>
  <button class="btn-outline rounded-l-none">Underline</button>
</div>

<!-- Segmented control (radio-based) -->
<div class="flex" role="group" aria-label="View mode">
  <input type="radio" name="view" id="list" class="sr-only" value="list" checked>
  <label for="list" class="btn-outline rounded-r-none border-r-0 cursor-pointer">List</label>
  <input type="radio" name="view" id="grid" class="sr-only" value="grid">
  <label for="grid" class="btn-outline rounded-l-none cursor-pointer">Grid</label>
</div>
```

---

## Input Group

No dedicated class — compose using Tailwind. Attach prefix/suffix to an input:

```html
<!-- Prefix text -->
<div class="flex">
  <span class="flex items-center px-3 rounded-l-md border border-r-0 border-input bg-muted text-sm text-muted-foreground">
    https://
  </span>
  <input class="input rounded-l-none" type="text" placeholder="example.com">
</div>

<!-- Suffix icon button -->
<div class="flex">
  <input class="input rounded-r-none" type="text" placeholder="Search...">
  <button class="btn-outline rounded-l-none border-l-0">
    <!-- search icon -->
  </button>
</div>

<!-- Both sides -->
<div class="flex">
  <span class="flex items-center px-3 rounded-l-md border border-r-0 border-input bg-muted text-sm">$</span>
  <input class="input rounded-none" type="number" placeholder="0.00">
  <span class="flex items-center px-3 rounded-r-md border border-l-0 border-input bg-muted text-sm">USD</span>
</div>
```

---

## Item

`item` is a generic list row component — used for settings lists, user lists, menus rendered in a card:

```html
<div class="item">
  <div class="item-indicator">
    <!-- optional: check icon, avatar, or icon -->
  </div>
  <div class="item-content">
    <span class="item-label">Item title</span>
    <span class="item-description">Secondary text</span>
  </div>
  <div class="item-actions">
    <!-- optional: badge, button, chevron -->
  </div>
</div>
```

Or more commonly as a `<button>` or `<a>` for interactive rows. In practice the docs show it used inside command menus and dropdowns.

---

## Empty

Empty state pattern — no dedicated class, uses Tailwind composition:

```html
<div class="flex flex-col items-center justify-center gap-4 py-16 text-center text-muted-foreground">
  <!-- optional icon -->
  <svg class="size-12 opacity-50" ...>...</svg>
  <div>
    <h3 class="font-semibold text-foreground">No results found</h3>
    <p class="text-sm mt-1">Try adjusting your search or filters.</p>
  </div>
  <button class="btn-outline">Reset filters</button>
</div>
```

---

## Command (Command Palette)

Command is a search-driven menu — uses the `.select` structure with a `role="listbox"` but driven by a search input. Requires JS (`command.min.js`).

```html
<div id="my-command" class="command">
  <div class="command-input-wrapper">
    <!-- search icon -->
    <input type="text" class="command-input" placeholder="Search commands..." />
  </div>
  <div class="command-list" role="listbox">
    <div role="group" aria-labelledby="group-calendar">
      <div role="heading" id="group-calendar">Calendar</div>
      <div role="option" data-value="new-event">
        <!-- icon -->
        New Event
        <kbd class="kbd">⌘N</kbd>
      </div>
    </div>
    <div role="group" aria-labelledby="group-settings">
      <div role="heading" id="group-settings">Settings</div>
      <div role="option" data-value="profile">Profile</div>
    </div>
  </div>
</div>
```

For a modal command palette, wrap in `<dialog class="dialog command-dialog">` and trigger with `.showModal()`.

Cherry-pick script: `<script src=".../js/command.min.js" defer></script>` (plus `basecoat.min.js`).

---

## Theme Switcher

No dedicated component — uses a small JS snippet in `<head>` plus a `basecoat:theme` custom event:

```html
<head>
  <!-- Prevents flash of unstyled content -->
  <script>
    (() => {
      const apply = (dark) => {
        document.documentElement.classList.toggle('dark', dark);
        try { localStorage.setItem('theme', dark ? 'dark' : 'light'); } catch (_) {}
      };
      try { apply(localStorage.getItem('theme') === 'dark'); } catch (_) {}
      document.addEventListener('basecoat:theme', (event) => {
        const mode = event.detail?.mode;
        apply(mode === 'dark' ? true : mode === 'light' ? false : !document.documentElement.classList.contains('dark'));
      });
    })();
  </script>
</head>

<!-- Toggle button dispatches the event -->
<button type="button"
  aria-label="Toggle dark mode"
  data-tooltip="Toggle dark mode"
  onclick="document.dispatchEvent(new CustomEvent('basecoat:theme'))"
  class="btn-icon-outline size-8">
  <!-- sun icon (shown in dark mode) -->
  <span class="hidden dark:block"><svg ...sun...</svg></span>
  <!-- moon icon (shown in light mode) -->
  <span class="block dark:hidden"><svg ...moon...</svg></span>
</button>

<!-- Force a specific mode -->
<button onclick="document.dispatchEvent(new CustomEvent('basecoat:theme', { detail: { mode: 'dark' } }))">
  Dark
</button>
```

---

## Themeing

Compatible with any shadcn/ui theme. To use a custom theme:

1. Go to ui.shadcn.com/themes, pick a theme, click "Copy code"
2. Save to `theme.css`
3. Import it: `@import "./theme.css";`

Theme variables control everything: `--color-primary`, `--color-card`, `--radius`, etc.

---

## htmx Integration Notes

**Dialogs + htmx:** Load dialog content dynamically but keep the `<dialog>` shell in the base template. Use htmx to fill `<section>` content, then call `.showModal()` via `HX-Trigger`:

```python
response['HX-Trigger'] = 'openDialog'
```
```html
<dialog id="edit-dialog" ... hx-on:openDialog="this.showModal()">
```

**Toast + htmx:** Best pattern — fire `basecoat:toast` via `HX-Trigger` response header after mutations. Zero front-end JS needed.

**Tabs + htmx:** Skip Basecoat JS tabs for server-driven content. Use htmx HATEOAS tab pattern where each tab response re-renders the full tab bar with active state. Use Basecoat's tab CSS classes for styling only.

**Skeleton + htmx:** Show skeleton on initial load, replace with real content via `hx-trigger="load"`:
```html
<div hx-get="/contacts/list" hx-trigger="load" hx-target="this" hx-swap="outerHTML">
  <div class="space-y-3">
    <div class="skeleton h-10 w-full"></div>
    <div class="skeleton h-10 w-full"></div>
    <div class="skeleton h-10 w-full"></div>
  </div>
</div>
```
