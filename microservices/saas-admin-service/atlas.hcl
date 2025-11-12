# Atlas configuration for SaaS Admin Service
# This configuration uses GORM models to generate database migrations

# Define the GORM models source
data "external_schema" "gorm" {
  program = [
    "go",
    "run",
    "-mod=mod",
    "ariga.io/atlas-provider-gorm",
    "load",
    "--path", "./internal/models",
    "--dialect", "postgres",
  ]
}

# Development environment
env "dev" {
  src = data.external_schema.gorm.url
  url = "postgres://postgres:postgres@localhost:5432/saas_admin?sslmode=disable"
  # Use a local dev database for schema comparisons
  dev = "postgres://postgres:postgres@localhost:5432/saas_admin_dev?sslmode=disable"

  migration {
    dir = "file://migrations"
  }

  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}

# Production environment
env "prod" {
  src = data.external_schema.gorm.url

  # Set via environment variable: ATLAS_URL
  # Example: postgres://user:pass@rds-endpoint:5432/saas_admin?sslmode=require
  url = getenv("ATLAS_URL")

  migration {
    dir = "file://migrations"
  }

  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }

  # Ensure migrations are applied linearly
  migration {
    dir = "file://migrations"
    revisions_schema = "atlas_schema_revisions"
  }
}
