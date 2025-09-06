# Enterprise Status Page

This is a comprehensive, enterprise-grade status page application built with Go and React. It provides real-time updates on service status, incident reporting, scheduled maintenance, and subscriber notifications.

## Features

- **Service Status**: Monitor the real-time status of all your services.
- **Incident Reporting**: Create and manage incidents with detailed descriptions and status updates.
- **Scheduled Maintenance**: Schedule and manage maintenance events.
- **Subscriber Notifications**: Allow users to subscribe for email notifications on incidents and maintenance.
- **Admin Dashboard**: A secure, JWT-protected dashboard for managing services, incidents, and maintenance.

## Getting Started

### Prerequisites

- [Go](https://golang.org/doc/install) (version 1.18 or higher)
- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/)

### Setup

1.  **Clone the repository:**

    ```sh
    git clone <repository-url>
    cd statuspage
    ```

2.  **Start the database:**

    This will start a PostgreSQL container in the background.

    ```sh
    docker-compose up -d
    ```

3.  **Run the application:**

    The application will connect to the database and start the server on port 8080.

    ```sh
    go run cmd/api/main.go
    ```

## How to Log In

1.  **Access the application:**

    Open your browser and navigate to `http://localhost:8080`.

2.  **Admin Login:**

    Navigate to the admin login page in your browser:

    -   **URL**: `http://localhost:8080/admin/login`

    Use the following credentials to log in:

    -   **Username**: `admin`
    -   **Password**: `password`

    Upon successful login, you will be redirected to the admin dashboard.

## API Endpoints

-   **Public Endpoints**:
    -   `GET /api/v1/status`: Get the status of all services.
    -   `GET /api/v1/incidents`: Get a list of all incidents.
    -   `GET /api/v1/maintenance`: Get a list of all scheduled maintenance events.
    -   `POST /api/v1/subscribers`: Subscribe to notifications.
    -   `DELETE /api/v1/subscribers`: Unsubscribe from notifications.
-   **Admin Endpoints** (require JWT authentication):
    -   `POST /api/v1/admin/incidents`: Create a new incident.
    -   `POST /api/v1/admin/maintenance`: Create a new maintenance event.
    -   ...and more.
