/**
 * Copyright 2024 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

locals {
  network = (
    var.vpc_create
    ? {
      id        = try(google_compute_network.network[0].id, null)
      name      = try(google_compute_network.network[0].name, null)
      self_link = try(google_compute_network.network[0].self_link, null)
    }
    : {
      id = format(
        "projects/%s/global/networks/%s",
        var.project_id,
        var.name
      )
      name = var.name
      self_link = format(
        "https://www.googleapis.com/compute/v1/projects/%s/global/networks/%s",
        var.project_id,
        var.name
      )
    }
  )
  peer_network = (
    var.peering_config == null
    ? null
    : element(reverse(split("/", var.peering_config.peer_vpc_self_link)), 0)
  )
}

resource "google_compute_network" "network" {
  count                                     = var.vpc_create ? 1 : 0
  project                                   = var.project_id
  name                                      = var.name
  description                               = var.description
  auto_create_subnetworks                   = var.auto_create_subnetworks
  delete_default_routes_on_create           = var.delete_default_routes_on_create
  mtu                                       = var.mtu
  routing_mode                              = var.routing_mode
  network_firewall_policy_enforcement_order = var.firewall_policy_enforcement_order
  enable_ula_internal_ipv6                  = var.ipv6_config.enable_ula_internal
  internal_ipv6_range                       = var.ipv6_config.internal_range
}

resource "google_compute_network_peering" "local" {
  provider             = google-beta
  count                = var.peering_config == null ? 0 : 1
  name                 = "${var.name}-${local.peer_network}"
  network              = local.network.self_link
  peer_network         = var.peering_config.peer_vpc_self_link
  export_custom_routes = var.peering_config.export_routes
  import_custom_routes = var.peering_config.import_routes
}

resource "google_compute_network_peering" "remote" {
  provider = google-beta
  count = (
    var.peering_config != null && try(var.peering_config.create_remote_peer, true)
    ? 1
    : 0
  )
  name                 = "${local.peer_network}-${var.name}"
  network              = var.peering_config.peer_vpc_self_link
  peer_network         = local.network.self_link
  export_custom_routes = var.peering_config.import_routes
  import_custom_routes = var.peering_config.export_routes
  depends_on           = [google_compute_network_peering.local]
}

resource "google_compute_shared_vpc_host_project" "shared_vpc_host" {
  provider   = google-beta
  count      = var.shared_vpc_host ? 1 : 0
  project    = var.project_id
  depends_on = [local.network]
}

resource "google_compute_shared_vpc_service_project" "service_projects" {
  provider = google-beta
  for_each = toset(
    var.shared_vpc_host && var.shared_vpc_service_projects != null
    ? var.shared_vpc_service_projects
    : []
  )
  host_project    = var.project_id
  service_project = each.value
  depends_on      = [google_compute_shared_vpc_host_project.shared_vpc_host]
}

resource "google_dns_policy" "default" {
  count                     = var.dns_policy == null ? 0 : 1
  project                   = var.project_id
  name                      = var.name
  enable_inbound_forwarding = try(var.dns_policy.inbound, null)
  enable_logging            = try(var.dns_policy.logging, null)
  networks {
    network_url = local.network.id
  }

  dynamic "alternative_name_server_config" {
    for_each = var.dns_policy.outbound != null ? [""] : []
    content {
      dynamic "target_name_servers" {
        for_each = (
          var.dns_policy.outbound.private_ns != null
          ? var.dns_policy.outbound.private_ns
          : []
        )
        iterator = ns
        content {
          ipv4_address    = ns.value
          forwarding_path = "private"
        }
      }
      dynamic "target_name_servers" {
        for_each = (
          var.dns_policy.outbound.public_ns != null
          ? var.dns_policy.outbound.public_ns
          : []
        )
        iterator = ns
        content {
          ipv4_address = ns.value
        }
      }
    }
  }
}
