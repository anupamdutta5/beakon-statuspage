# PostgreSQL Initialization Scripts

Place SQL scripts here to be executed when PostgreSQL container first starts.

Scripts are executed in alphabetical order.

Example:
```
01_create_databases.sql
02_create_users.sql
03_grant_permissions.sql
```

Note: These scripts only run on first container startup (when data volume is empty).
