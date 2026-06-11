resource "epilayer_instance" "target" {
  # ...
}

resource "epilayer_snapshot" "example" {
  name        = "example"
  instance_id = epilayer_instance.target.id

  retain_on_delete = true # optional
}
