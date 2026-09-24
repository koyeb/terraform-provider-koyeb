# Snapshots are taken from an existing volume and can later be
# used to create new volumes via koyeb_volume.snapshot_id.
resource "koyeb_volume" "my-volume" {
  name     = "my-volume"
  max_size = 10
  region   = "was"
}

resource "koyeb_snapshot" "my-snapshot" {
  name             = "my-snapshot"
  parent_volume_id = koyeb_volume.my-volume.id
}
