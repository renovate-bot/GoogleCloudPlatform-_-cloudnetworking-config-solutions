folder_id                 = ""
organization_id           = ""
bootstrap_project_id      = ""
network_hostproject_id    = ""
network_serviceproject_id = "" // <service(producer/consumer)-project-id>

organization_administrator = ["user:user-example@example.com"]
networking_administrator   = ["user:user-example@example.com"]
security_administrator     = ["user:user-example@example.com"]

producer_cloudsql_administrator = ["user:user-example@example.com"]
producer_gke_administrator      = ["user:user-example@example.com"]
producer_alloydb_administrator  = ["user:user-example@example.com"]
producer_vertex_administrator   = ["user:user-example@example.com"]
producer_mrc_administrator      = ["user:user-example@example.com"]

producer_connectivity_administrator = ["user:user-example@example.com"]

consumer_gce_administrator           = ["user:user-example@example.com"]
consumer_cloudrun_administrator      = ["user:user-example@example.com"]
consumer_workbench_administrator     = ["user:workbench-user-example@example.com"]
consumer_mig_administrator           = ["user:mig-user-example@example.com"]
consumer_umig_administrator          = ["user:umig-user-example@example.com"]
consumer_vpc_connector_administrator = ["user:user-example@example.com"]
consumer_appengine_administrator     = ["user:user-example@example.com"]
consumer_load_balacing_administrator = ["user:user-example@example.com"]
nsi_administrator                    = ["user:nsi-user-example@example.com"]
