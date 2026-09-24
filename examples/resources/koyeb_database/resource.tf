# A database is deployed as a service of type DATABASE inside an app
# named after the database. Deleting the database deletes the service
# but leaves the app in place.
resource "koyeb_database" "my-database" {
  name          = "my-database"
  pg_version    = 16
  region        = "was"
  instance_type = "free"
  db_name       = "koyebdb"
  db_owner      = "koyeb-adm"
}
