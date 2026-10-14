data "epilayer_images" "cloud-images" {
  filter = {
    type = "cloud-image"
  }
}

data "epilayer_images" "snapshots" {
  filter = {
    type = "snapshot"
  }
}

data "epilayer_images" "preconfigured-images" {
  filter = {
    type = "preconfigured"
  }
}
