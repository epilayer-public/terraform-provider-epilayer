data "epilayer_images" "cloud-images" {
  filter = {
    type = "cloud-image"
  }
}

data "epilayer_images" "snapshots" {
  filter = {
    type   = "snapshot"
    region = "NORD-NO-KRS-1"
  }
}

data "epilayer_images" "preconfigured-images" {
  filter = {
    type = "preconfigured"
  }
}
