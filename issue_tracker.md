# Issue Tracker

This file tracks the issues, fixes, and improvements for the enterprise status page application.

## 2025-08-30

### Initial Project Setup

*   **Action:** Set up the initial project structure for the Go application.
*   **Reason:** To establish a clean and scalable architecture from the beginning.
*   **Prevention:** Following best practices for Go project layouts will help prevent organizational issues in the future.

### JWT Authentication for Admin Routes

*   **Action:** Implemented JWT-based authentication for all admin-level API endpoints.
*   **Reason:** To secure the application and prevent unauthorized access to sensitive operations like creating incidents or updating service statuses.
*   **Prevention:** A centralized authentication middleware was created. All future admin routes should be placed within the authenticated route group to ensure they are protected by default.

### Scheduled Maintenance Management

*   **Action:** Added functionality to create, view, update, and delete scheduled maintenance events.
*   **Reason:** To proactively inform users about planned downtime or service interruptions, which is a core feature of an enterprise status page.
*   **Prevention:** The feature was built with a dedicated service and model, following the existing architecture. This modular approach should be continued for new features to maintain separation of concerns.

### Monitor API Handler Bug Fix

*   **Issue**: The `updateMonitor` API handler had a bug causing a compilation error (`assignment mismatch`) and incorrect logic. It fetched a monitor but immediately overwrote it instead of updating it.
*   **Fix**: Corrected the handler to properly receive both return values from `GetMonitorByID`. The logic was fixed to bind the incoming JSON data onto the fetched monitor object before saving.
*   **Prevention**: Ensure return values from functions are always checked and handled correctly. When updating an existing resource, always fetch it first and then apply changes to the fetched object to avoid data loss.

### Implemented Component Subscription Feature

*   **Issue:** The application lacked the ability for users to subscribe to notifications for specific services (components).
*   **Fix:** 
    - **Frontend**: Added a subscription modal to `index.html` with a dynamic checklist of services fetched from the backend. Updated `app.js` to handle modal interactions and form submission. Added styles for the modal in `style.css`.
    - **Backend**: Updated the `subscribe` handler in `internal/api/server.go` to accept a list of service IDs. Verified that `subscriberService.Subscribe` in `internal/services/subscriber_service.go` correctly associates subscribers with services.
    - **Notifications**: Confirmed that `notificationService` in `internal/services/notification_service.go` correctly sends notifications only to subscribers of affected services.
*   **Resolution**: The component subscription feature is now fully implemented, allowing users to receive targeted notifications.
*   **Enhancements**:
    - **Frontend Feedback**: The subscription modal in `app.js` was updated to provide UI feedback, including disabling the submit button during requests and displaying success or error messages.
    - **Backend Validation**: Added validation in `subscriber_service.go` to ensure all service IDs are valid before creating a subscription. The API handler in `server.go` was updated to return a `400 Bad Request` for invalid IDs.

### UI Modernization

*   **Issue:** The application's frontend had an outdated design.
*   **Fix:** Refactored the global stylesheet and updated all HTML templates (`index.html`, `dashboard.html`, `login.html`) to use a modern, card-based design with a consistent color scheme and typography.
*   **Prevention:** Adhere to the established design system in `style.css` for all new UI components to maintain a consistent and professional look.

### Web-Based Admin Login

*   **Action:** Implemented a web-based login page and dashboard for administrators. The authentication middleware was updated to handle both cookie-based sessions for the web UI and JWT tokens for the API.
*   **Reason:** To improve usability by removing the need for an API client to log in, providing a more traditional and user-friendly admin experience.
*   **Prevention:** The authentication middleware now centrally handles both authentication methods. New protected web routes should be added to the authenticated web group to ensure they are covered by the cookie-based auth check.

### Admin Logout Feature

*   **What was done:** Added a logout feature to the admin dashboard. This includes a logout link in the UI and a backend route that clears the authentication cookie.
*   **Why it was done:** To allow administrators to securely log out of their sessions.
*   **Future prevention:** N/A

### Maintenance Management Compilation and Runtime Fixes

*   **Date:** 2025-08-30
*   **Status:** Fixed
*   **Issue:** The initial implementation of the maintenance management feature had several compilation and runtime errors due to mismatches between the handlers, services, and models.
*   **Fixes:**
  1.  **Missing Service Method:** Added the `GetMaintenanceEventByID` method to `internal/services/maintenance_service.go` to allow fetching a single maintenance event.
  2.  **Incorrect Model Fields:** Corrected the field names in `internal/api/server.go` from `ScheduledAt` to `StartAt` to match the `models.Maintenance` struct.
  3.  **Type Mismatches:** Implemented string-to-uint conversion for maintenance event IDs passed in URL parameters.
  4.  **Incorrect Handler Logic:** Rewrote the maintenance handlers (`handleNewMaintenance`, `showEditMaintenancePage`, `handleEditMaintenance`, `handleDeleteMaintenance`) to correctly parse form data, call the appropriate service methods with the correct arguments, and handle errors.
  5.  **Template Errors:** Updated `web/templates/maintenance_form.html` to use the correct form field names (`start_at` and `end_at`) to ensure data binding with the backend.
*   **Prevention:** When implementing new features, ensure end-to-end consistency between the model, service, and handler layers. Write unit or integration tests to catch these mismatches early.
