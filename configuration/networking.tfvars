project_id = ""
region     = ""

## VPC input variables

network_name = ""
subnets = [
  {
    name                  = ""
    ip_cidr_range         = ""
    region                = ""
    enable_private_access = false # Use true or false
  }
]

# Configuration for setting up a Shared VPC Host project, enabling centralized network management and resource sharing across multiple projects.
shared_vpc_host = false

# PSC/Service Connecitvity Variables

create_scp_policy      = false # Use true or false based on your requirements
subnets_for_scp_policy = [""]  # List subnets here from the same region as the SCP

## Cloud Nat input variables
create_nat = false # Use true or false 

## Cloud HA VPN input variables

create_havpn = false
peer_gateways = {
  default = {
    gcp = "" # e.g. projects/<google-cloud-peer-projectid>/regions/<google-cloud-region>/vpnGateways/<peer-vpn-name>
  }
}

tunnel_1_router_bgp_session_range = ""
tunnel_1_bgp_peer_asn             = 64514
tunnel_1_bgp_peer_ip_address      = ""
tunnel_1_shared_secret            = ""

tunnel_2_router_bgp_session_range = ""
tunnel_2_bgp_peer_asn             = 64514
tunnel_2_bgp_peer_ip_address      = ""
tunnel_2_shared_secret            = ""

## Cloud Interconnect input variables

create_interconnect = false # Use true or false