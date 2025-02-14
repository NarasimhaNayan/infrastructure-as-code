output "vnet_id" {
  value = azurerm_virtual_network.virtual_network.id
}

output "public_subnet_id" {
  value = azurerm_subnet.public_subnet.id
}

output "private_subnet_id" {
  value = azurerm_subnet.private_subnet.id
}
