output "resource_group_name" {
  value = azurerm_resource_group.resource_group.name
}

output "virtual_network_id" {
  value = module.network.vnet_id
}

output "private_subnet_id" {
  value = module.network.private_subnet_id
}

output "aks_cluster_name" {
  value = azurerm_kubernetes_cluster.aks.name
}

# output "postgres_server_fqdn" {
#   value = azurerm_postgresql_server.postgres.fqdn
# }

# output "storage_account_endpoint" {
#   value = azurerm_storage_account.storage.primary_web_endpoint
# }
