# htmx UI Examples Reference

All examples from https://htmx.org/examples/ — canonical HTML patterns.

---

## Click to Edit
Inline editing without page reload. `hx-target="this"` + `hx-swap="outerHTML"` swaps the view div with an edit form and back.

**View state:**
```html
<div hx-target="this" hx-swap="outerHTML">
  <div><label>First Name</label>: Joe</div>
  <div><label>Last Name</label>: Blow</div>
  <div><label>Email</label>: joe@blow.com</div>
  <button hx-get="/contact/1/edit" class="btn primary">Click To Edit</button>
</div>
```

**Edit state (returned by server):**
```html
<form hx-put="/contact/1" hx-target="this" hx-swap="outerHTML">
  <div><label>First Name</label><input type="text" name="firstName" value="Joe"></div>
  <div><label>Last Name</label><input type="text" name="lastName" value="Blow"></div>
  <div><label>Email</label><input type="email" name="email" value="joe@blow.com"></div>
  <button type="submit">Submit</button>
  <button hx-get="/contact/1">Cancel</button>
</form>
```

---

## Bulk Update
Wrap table in a form. Checkboxes manage their own state — no need to re-render the table.

```html
<form id="checked-contacts" hx-post="/users" hx-swap="innerHTML settle:3s" hx-target="#toast">
  <table>
    <tbody id="tbody">
      <tr>
        <td>Joe Smith</td>
        <td>joe@smith.org</td>
        <td><input type="checkbox" name="active:joe@smith.org"></td>
      </tr>
    </tbody>
  </table>
  <input type="submit" value="Bulk Update" class="btn primary">
  <output id="toast"></output>
</form>
```

CSS fade-out toast:
```css
#toast { background: #E1F0DA; opacity: 0; transition: opacity 3s ease-out; }
#toast.htmx-settling { opacity: 100; }
```

---

## Click to Load (Pagination)
Replace the "load more" row with the next page of results (which itself contains the next "load more" row).

```html
<tr id="replaceMe">
  <td colspan="3">
    <button class="btn primary"
            hx-get="/contacts/?page=2"
            hx-target="#replaceMe"
            hx-swap="outerHTML">
      Load More Agents... <img class="htmx-indicator" src="/img/bars.svg" alt="">
    </button>
  </td>
</tr>
```

---

## Delete Row
Inherit `hx-target` and `hx-confirm` from `<tbody>` so all delete buttons share the same config.

```html
<tbody hx-confirm="Are you sure?" hx-target="closest tr" hx-swap="outerHTML swap:1s">
  <tr>
    <td>Angie MacDowell</td>
    <td>angie@macdowell.org</td>
    <td>Active</td>
    <td><button class="btn danger" hx-delete="/contact/1">Delete</button></td>
  </tr>
</tbody>
```

Fade-out animation via CSS class applied during swap:
```css
tr.htmx-swapping td { opacity: 0; transition: opacity 1s ease-out; }
```

Server returns `200` with empty body to remove the row.

---

## Edit Row
Editable table row using `hx-include` (since `<form>` can't go inside `<tr>`).

**Read-only row:**
```html
<tr>
  <td>Joe Smith</td>
  <td>joe@smith.org</td>
  <td><button hx-get="/contact/1/edit" hx-trigger="edit" onClick="htmx.trigger(this, 'edit')">Edit</button></td>
</tr>
```

**Edit row (returned by server):**
```html
<tr hx-trigger="cancel" class="editing" hx-get="/contact/1">
  <td><input name="name" value="Joe Smith" autofocus></td>
  <td><input name="email" value="joe@smith.org"></td>
  <td>
    <button hx-get="/contact/1">Cancel</button>
    <button hx-put="/contact/1" hx-include="closest tr">Save</button>
  </td>
</tr>
```

Note: `hx-include="closest tr"` collects all inputs in the row since forms can't wrap `<tr>` elements.

---

## Lazy Loading
Show a spinner on page load, replace with actual content when loaded.

```html
<div hx-get="/graph" hx-trigger="load">
  <img alt="Loading..." class="htmx-indicator" width="150" src="/img/bars.svg"/>
</div>
```

Fade-in CSS:
```css
.htmx-settling img { opacity: 0; }
img { transition: opacity 300ms ease-in; }
```

---

## Inline Validation
Per-field validation that replaces just that field's wrapper div.

```html
<form hx-post="/contact">
  <div hx-target="this" hx-swap="outerHTML">
    <label>Email Address</label>
    <input name="email" hx-post="/contact/email" hx-indicator="#ind">
    <img id="ind" src="/img/bars.svg" class="htmx-indicator"/>
  </div>
  <div><label>First Name</label><input name="firstName"></div>
  <button class="btn primary">Submit</button>
</form>
```

Server returns the wrapper div annotated with `.error` or `.valid`:
```html
<div hx-target="this" hx-swap="outerHTML" class="error">
  <label>Email Address</label>
  <input name="email" hx-post="/contact/email" value="bad@email">
  <div class="error-message">That email is already taken.</div>
</div>
```

```css
.error input { box-shadow: 0 0 3px #CC0000; }
.valid input { box-shadow: 0 0 3px #36cc00; }
```

---

## Infinite Scroll
Last row triggers load when scrolled into view. The loaded content's last row does the same.

```html
<tr hx-get="/contacts/?page=2"
    hx-trigger="revealed"
    hx-swap="afterend">
  <td>Agent Smith</td>
  <td>void29@null.org</td>
</tr>
```

> Use `hx-trigger="intersect once"` instead of `revealed` if you're using `overflow-y: scroll`.

---

## Active Search
Debounced live search. Multiple triggers separated by commas.

```html
<input type="search"
       name="search"
       placeholder="Search..."
       hx-post="/search"
       hx-trigger="input changed delay:500ms, keyup[key=='Enter'], load"
       hx-target="#search-results"
       hx-indicator=".htmx-indicator">

<table>
  <tbody id="search-results"></tbody>
</table>
```

- `delay:500ms` — wait until user stops typing
- `changed` — don't fire if value didn't change (e.g. arrow keys)
- `load` — show all results on initial page load
- `keyup[key=='Enter']` — also trigger on Enter

---

## Progress Bar
Polling-based progress bar. Polling stops when the server returns HTML without `hx-trigger`.

**Start button:**
```html
<div hx-target="this" hx-swap="outerHTML">
  <button hx-post="/start">Start Job</button>
</div>
```

**Running state (server returns this after POST /start):**
```html
<div hx-trigger="done" hx-get="/job" hx-swap="outerHTML" hx-target="this">
  <h3 id="pblabel">Running</h3>
  <div hx-get="/job/progress"
       hx-trigger="every 600ms"
       hx-target="this"
       hx-swap="innerHTML">
    <div class="progress" role="progressbar">
      <div id="pb" class="progress-bar" style="width:0%"></div>
    </div>
  </div>
</div>
```

**Complete: server sends `HX-Trigger: done` response header** which triggers the outer `hx-trigger="done"` to reload the whole div as "Complete" state.

```css
.progress-bar { transition: width .6s ease; }
```

---

## Cascading Selects
Second select updates when first changes.

```html
<select name="make" hx-get="/models" hx-target="#models" hx-indicator=".htmx-indicator">
  <option value="audi">Audi</option>
  <option value="toyota">Toyota</option>
</select>

<select id="models" name="model">
  <option value="a1">A1</option>
</select>
```

Server returns just the `<option>` elements for the selected make.

---

## File Upload with Progress

```html
<form id="upload-form" hx-encoding="multipart/form-data" hx-post="/upload">
  <input type="file" name="file">
  <button>Upload</button>
  <progress id="progress" value="0" max="100"></progress>
</form>
<script>
  htmx.on('#upload-form', 'htmx:xhr:progress', function(evt) {
    htmx.find('#progress').setAttribute('value', evt.detail.loaded/evt.detail.total * 100)
  });
</script>
```

---

## Updating Other Content
Four approaches when a response needs to update elements outside the trigger's `hx-target`.

### Option 1: Expand the target (simplest)
Wrap both elements, target the wrapper.
```html
<div id="table-and-form">
  <table>...</table>
  <form hx-post="/contacts" hx-target="#table-and-form">...</form>
</div>
```
Server re-renders both table and form.

### Option 2: Out-of-band swap (OOB)
Server response includes extra elements with `hx-swap-oob`:
```html
<!-- Primary response (replaces form target) -->
<form>...</form>

<!-- OOB: appended to contacts table regardless of form's hx-target -->
<tbody hx-swap-oob="beforeend:#contacts-table">
  <tr><td>Joe Smith</td><td>joe@smith.com</td></tr>
</tbody>
```

### Option 3: Server-sent event trigger
Table listens for a custom event fired by a response header.

```html
<tbody id="contacts-table" hx-get="/contacts/table" hx-trigger="newContact from:body">
```

Server responds with header: `HX-Trigger: newContact`

This triggers the table to refresh itself independently.

### Option 4: Path-deps extension
```html
<tbody hx-get="/contacts/table"
       hx-ext="path-deps"
       hx-trigger="path-deps"
       path-deps="/contacts">
```
Auto-refreshes when any request hits `/contacts`.

**Recommendation:** Use Option 1 (expand target) when elements are close in the DOM. Option 2 (OOB) or Option 3 (events) for elements far apart. Option 3 is cleaner for event-driven architectures.

---

## Tabs (HATEOAS style)
Active tab is encoded in the server response — no client state needed.

**Initial load:**
```html
<div id="tabs" hx-get="/tab1" hx-trigger="load delay:100ms" hx-target="#tabs" hx-swap="innerHTML"></div>
```

**Each tab response includes the full tab bar** with the active tab marked:
```html
<div class="tab-list">
  <button hx-get="/tab1" class="selected">Tab 1</button>
  <button hx-get="/tab2">Tab 2</button>
  <button hx-get="/tab3">Tab 3</button>
</div>
<div id="tab-content">Tab 1 content here...</div>
```

---

## Custom Modal Dialog
Load modal content from server, append to `<body>`, use Hyperscript for close animation.

**Trigger button:**
```html
<button hx-get="/modal" hx-target="body" hx-swap="beforeend">Open Modal</button>
```

**Server returns modal fragment:**
```html
<div id="modal" _="on closeModal add .closing then wait for animationend then remove me">
  <div class="modal-underlay" _="on click trigger closeModal"></div>
  <div class="modal-content">
    <h1>Modal Dialog</h1>
    Content here.
    <button _="on click trigger closeModal">Close</button>
  </div>
</div>
```

Key CSS: position `fixed`, animate with `fadeIn`/`zoomIn` keyframes, add `.closing` class to trigger `fadeOut`/`zoomOut` before removal.

---

## Sortable (Drag & Drop)
Integrate Sortable.js via htmx event hooks. POST new order on drag end.

```html
<form class="sortable" hx-post="/items" hx-trigger="end">
  <div class="htmx-indicator">Updating...</div>
  <div><input type="hidden" name="item" value="1"/>Item 1</div>
  <div><input type="hidden" name="item" value="2"/>Item 2</div>
  <div><input type="hidden" name="item" value="3"/>Item 3</div>
</form>

<script>
htmx.onLoad(function(content) {
  var sortables = content.querySelectorAll(".sortable");
  sortables.forEach(function(sortable) {
    var sortableInstance = new Sortable(sortable, {
      animation: 150,
      onEnd: function(evt) { this.option("disabled", true); }
    });
    sortable.addEventListener("htmx:afterSwap", function() {
      sortableInstance.option("disabled", false);
    });
  });
});
</script>
```

---

## Keyboard Shortcuts
Add keyboard triggers alongside click using `from:body` to make them global.

```html
<button hx-trigger="click, keyup[altKey&&shiftKey&&key=='D'] from:body"
        hx-post="/doit">
  Do It! (alt-shift-D)
</button>
```

---

## Custom Confirmation Dialog
Replace the browser's native `confirm()` with a custom dialog (e.g. SweetAlert2).

**Using `htmx:confirm` event globally:**
```html
<script>
document.addEventListener("htmx:confirm", function(e) {
  if (!e.detail.question) return;
  e.preventDefault();
  Swal.fire({ title: "Proceed?", text: e.detail.question })
    .then(function(result) {
      if (result.isConfirmed) e.detail.issueRequest(true);
    });
});
</script>

<!-- Works on any element with hx-confirm -->
<button hx-delete="/contact/1" hx-confirm="Delete this contact?">Delete</button>

<!-- Dynamic confirm text in a loop (Django example) -->
{% for client in clients %}
<button hx-post="/delete/{{client.pk}}" hx-confirm="Delete {{client.name}}?">Delete</button>
{% endfor %}
```

---

## Async Authentication
Delay htmx requests until an auth token is available.

```html
<script>
  let authToken = null;
  auth.then((token) => { authToken = token; });

  // Block requests until token is ready
  htmx.on("htmx:confirm", (e) => {
    if (authToken == null) {
      e.preventDefault();
      auth.then(() => e.detail.issueRequest());
    }
  });

  // Inject token into all requests
  htmx.on("htmx:configRequest", (e) => {
    e.detail.headers["AUTH"] = authToken;
  });
</script>

<button hx-post="/example" hx-target="next output">Authenticated Request</button>
```

---

## hx-swap-oob Quick Reference

`hx-swap-oob` allows server responses to update elements *outside* the primary `hx-target`:

```html
<!-- Primary content (replaces hx-target) -->
<div>Main response content</div>

<!-- OOB: updates #notification regardless of where the request came from -->
<div id="notification" hx-swap-oob="true">You have 3 new messages</div>

<!-- OOB with specific swap strategy -->
<tbody hx-swap-oob="beforeend:#contacts-table">
  <tr><td>New Row</td></tr>
</tbody>
```

---

## HX-Trigger Response Header Quick Reference

Server can fire client-side events via response headers:

```python
# Django example
from django.http import HttpResponse

response = HttpResponse(content)
response['HX-Trigger'] = 'contactSaved'           # single event
response['HX-Trigger'] = 'contactSaved, tableUpdated'  # multiple events
response['HX-Trigger-After-Swap'] = 'closeModal'  # fires after swap completes
response['HX-Trigger-After-Settle'] = 'done'      # fires after CSS settle
```

Elements listen for these:
```html
<div hx-get="/refresh" hx-trigger="contactSaved from:body">...</div>
```
