# Atlas configuration for Tenant Admin Service
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
  url = "postgres://postgres:postgres@localhost:5432/tenant_admin_db?sslmode=disable"
  dev = "docker://postgres/16/dev?search_path=public"

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
  # Example: postgres://user:pass@rds-endpoint:5432/tenant_admin_db?sslmode=require
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
