terraform {
  required_version = ">= 0.14"
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~>3.0"
    }
  }
}

provider "azurerm" {
  features {}
}
#   To connect Azure and Terraform
#   https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/guides/service_principal_client_secret


# Create a resource group for all resources
resource "azurerm_resource_group" "resource_group" {
  name     = var.resource_group_name
  location = var.location
}

# # Invoke the custom network module to create VNet, subnets, and NSGs
module "network" {
  source              = "./modules/network"
  resource_group_name = azurerm_resource_group.resource_group.name
  location            = azurerm_resource_group.resource_group.location
  virtual_network_name= var.virtual_network_name
  address_space       = var.vnet_address_space
  public_subnet_prefix  = var.public_subnet_prefix
  private_subnet_prefix = var.private_subnet_prefix
}

# # Provision an AKS cluster using the private subnet
resource "azurerm_kubernetes_cluster" "aks" {
  name                = var.aks_cluster_name
  location            = azurerm_resource_group.resource_group.location
  resource_group_name = azurerm_resource_group.resource_group.name
  dns_prefix          = var.aks_dns_prefix

  default_node_pool {
    name            = "agentpool"
    node_count      = var.aks_node_count
    vm_size         = var.aks_vm_size
    vnet_subnet_id  = module.network.private_subnet_id
  }

  identity {
    type = "SystemAssigned"
  }

  network_profile {
    network_plugin = "azure"
    network_policy = "azure"
    service_cidr = "10.1.0.0/16"   # New, non-overlapping service CIDR
    dns_service_ip = "10.1.0.10"     # An IP within the new service CIDR
  }

  tags = {
    Environment = "Dev"
  }
}

# Create a managed PostgreSQL server (managed database service)
resource "azurerm_postgresql_server" "postgres" {
  name                = var.postgres_server_name
  location            = azurerm_resource_group.resource_group.location
  resource_group_name = azurerm_resource_group.resource_group.name
  administrator_login = var.postgres_admin
  administrator_login_password = var.postgres_admin_password

  sku_name   = "B_Gen5_1"
  version    = "11"
  storage_mb = 5120

  backup_retention_days      = 7
  geo_redundant_backup_enabled = false
  auto_grow_enabled          = true

  # Enforce secure connections
  ssl_enforcement_enabled = true

  tags = {
    Environment = "Dev"
  }
}

# Create a PostgreSQL database on the server
resource "azurerm_postgresql_database" "database" {
  name                = var.postgres_database_name
  resource_group_name = azurerm_resource_group.resource_group.name
  server_name         = azurerm_postgresql_server.postgres.name
  charset             = "UTF8"
  collation           = "English_United States.1252"
}

# Provision an Azure Storage Account for data storage
resource "azurerm_storage_account" "storage" {
  name                     = var.storage_account_name
  resource_group_name      = azurerm_resource_group.resource_group.name
  location                 = azurerm_resource_group.resource_group.location
  account_tier             = "Standard"
  account_replication_type = "LRS"

  # Enforce HTTPS and TLS best practices
  https_traffic_only_enabled = true
  min_tls_version          = "TLS1_2"

  tags = {
    Environment = "Dev"
  }
}

