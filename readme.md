# Hacker News Clone (Gin + PostgreSQL)

This is a simple Hacker News clone built with Gin and PostgreSQL. The purpose of this project is to educate users on how to deploy database-dependent applications on Leapcell.

## Features

- Gin web framework for backend
- PostgreSQL database integration
- HTML templating for rendering views
- Static files support

## Project Structure

```
.
├── main.go               # Main application entry point
├── go.mod                # Go module file
├── go.sum                # Go dependencies file
├── static/               # Directory for static assets (CSS, JavaScript, images)
└── templates/            # HTML templates for rendering views
    ├── index.html        # Homepage displaying posts
    ├── post_detail.html  # Template for displaying post details
```

## Deployment on Leapcell

This guide will walk you through setting up and deploying the project on Leapcell.

### Prerequisites

Ensure you have the following:

- A Leapcell account
- PostgreSQL database instance
- Go installed (recommended: Go 1.18+)

### Environment Variables

This project requires a PostgreSQL connection string, which should be set using the following environment variable:

```bash
PG_DSN=<your_postgresql_connection_string>
```

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/leapcell/gin-hackernews-clone
   cd gin-hackernews-clone
   ```
2. Install dependencies:
   ```bash
   go mod tidy
   ```

### Running Locally

To start the project locally, ensure your PostgreSQL instance is running and execute:

```bash
go run main.go
```

The application will be accessible at `http://localhost:8080`.

### Deploying on Leapcell

1. Push your code to a GitHub repository.
2. Log in to Leapcell and connect your repository.
3. Configure the `PG_DSN` environment variable in the Leapcell deployment settings.
4. Deploy your application.

Once deployed, your application will be accessible via the Leapcell-generated domain.

## Contributing

Feel free to submit issues or pull requests to improve this project.

## Contact

For support, reach out via the Leapcell Discord community or email support@leapcell.io.
