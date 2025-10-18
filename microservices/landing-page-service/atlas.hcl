# Atlas configuration for Landing Page Service

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

  # Development database URL
  url = "postgres://postgres:postgres@localhost:5432/statuspage_landing?sslmode=disable"

  # Development database for schema diffing
  dev = "docker://postgres/16/dev?search_path=public"

  # Migration directory
  migration {
    dir = "file://migrations"
  }

  # Format migrations with proper SQL formatting
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}

# Production environment (for future use)
env "prod" {
  src = data.external_schema.gorm.url

  # Production database URL (from environment variable)
  url = getenv("DATABASE_URL")

  migration {
    dir = "file://migrations"
  }

  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}
