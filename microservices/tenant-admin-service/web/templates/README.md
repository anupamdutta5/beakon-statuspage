# Template Structure

This directory contains the modular template files for the Tenant Admin Service web interface.

## Directory Structure

```
templates/
├── admin_dashboard.html     # Main dashboard page (uses components)
├── login.html              # Login page
├── components/             # Reusable template components
│   ├── header.html        # Top bar with page title and user info
│   ├── sidebar.html       # Left navigation sidebar
│   ├── scripts.html       # JavaScript functions
│   └── styles.html        # CSS styles
└── layouts/               # Base layouts (reserved for future use)
```

## Component Usage

### Main Dashboard
The `admin_dashboard.html` file imports and uses all components:

```go
{{template "styles" .}}    // CSS styles
{{template "sidebar" .}}   // Navigation sidebar
{{template "header" .}}    // Top bar header
{{template "scripts" .}}   // JavaScript functions
```

### Components

1. **styles.html** - Contains all CSS styles for:
   - Sidebar navigation
   - Header/top bar
   - Stats cards
   - Action cards
   - Empty states
   - Responsive layouts

2. **sidebar.html** - Navigation sidebar with:
   - Tenant logo and name
   - Main navigation (Dashboard, Status Pages, Components, Incidents)
   - Manage section (Subscribers, Settings)
   - Logout button

3. **header.html** - Top bar displaying:
   - Current page title
   - User avatar
   - Tenant name
   - Administrator role

4. **scripts.html** - JavaScript functions:
   - `showSection()` - Navigate between sections
   - `logout()` - Handle user logout

## Benefits of Modular Structure

1. **Easy Maintenance** - Update styles, sidebar, or scripts in one place
2. **Code Reusability** - Components can be reused across multiple pages
3. **Better Organization** - Clear separation of concerns
4. **Faster Development** - Add new pages by composing existing components
5. **Consistency** - Shared components ensure consistent UI/UX

## Adding New Pages

To create a new page using the modular components:

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>{{.title}}</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/css/bootstrap.min.css" rel="stylesheet">
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/css/all.min.css" rel="stylesheet">
    {{template "styles" .}}
</head>
<body>
    {{template "sidebar" .}}
    <div class="main-content">
        {{template "header" .}}
        <div class="content-area">
            <!-- Your page content here -->
        </div>
    </div>
    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/js/bootstrap.bundle.min.js"></script>
    {{template "scripts" .}}
</body>
</html>
```

## Modifying Components

### To update the sidebar navigation:
Edit `components/sidebar.html`

### To update styles:
Edit `components/styles.html`

### To add new JavaScript functions:
Edit `components/scripts.html`

### To modify the header:
Edit `components/header.html`

## Notes

- All components use Go template syntax (`{{template "name" .}}`)
- The `.` passes the current context data to the component
- Components are defined using `{{define "name"}}...{{end}}`
- Bootstrap 5.1.3 and Font Awesome 6.0.0 are included by default
